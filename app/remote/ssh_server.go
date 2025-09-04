package remote

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/app/help"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func RunSSHServer() {
	sshCfg, err := prepareConfig()
	if err != nil {
		log.Fatal(help.PrintErr(err))
	}
	if err := serve(sshCfg); err != nil {
		log.Fatal(help.PrintErr(err))
	}
}

func serve(sshCfg *ssh.ServerConfig) error {
	cfg := conf.GetConf()

	connStr := fmt.Sprintf("0.0.0.0:%d", cfg.SSHPort)
	fmt.Printf("Listening on: %q\n", connStr)
	listener, err := net.Listen("tcp", connStr)
	if err != nil {
		return help.WrapErr(err, "listening for connection")
	}
	fmt.Println("1")
	nConn, err := listener.Accept()
	if err != nil {
		return help.WrapErr(err, "accepting incoming connection")
	}
	fmt.Println("2")

	conn, chans, reqs, err := ssh.NewServerConn(nConn, sshCfg)
	if err != nil {
		fmt.Println("3")
		return help.WrapErr(err, "handshaking conn")
	}
	fmt.Println("4")
	log.Printf("logged in with key %s", conn.Permissions.Extensions["pubkey-fp"])

	var wg sync.WaitGroup
	defer wg.Wait()

	// The incoming Request channel must be serviced.
	wg.Add(1)
	go func() {
		ssh.DiscardRequests(reqs)
		wg.Done()
	}()

	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			return help.WrapErr(err, "accepting channel")
		}

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "shell" request.
		wg.Add(1)
		go func(in <-chan *ssh.Request) {
			for req := range in {
				req.Reply(req.Type == "shell", nil)
			}
			wg.Done()
		}(requests)

		term := term.NewTerminal(channel, "> ")

		wg.Add(1)
		go func() {
			defer func() {
				channel.Close()
				wg.Done()
			}()
			for {
				line, err := term.ReadLine()
				if err != nil {
					break
				}
				fmt.Println(line)
			}
		}()
	}

	return nil
}

func prepareConfig() (*ssh.ServerConfig, error) {
	fmt.Println("Starting SSH Server...")
	cfg := conf.GetConf()
	fmt.Printf("Home directory: %q\n", cfg.MaeveDir)

	authorizedKeysMap := map[string]bool{}
	for _, key := range cfg.SSHAuthKeys {
		pubKey, err := ssh.ParsePublicKey([]byte(key))
		if err != nil {
			return nil, help.WrapErr(err, "parsing public key")
		}

		authorizedKeysMap[string(pubKey.Marshal())] = true
	}

	algorithms := ssh.SupportedAlgorithms()
	config := &ssh.ServerConfig{
		Config: ssh.Config{
			KeyExchanges: algorithms.KeyExchanges,
			Ciphers:      algorithms.Ciphers,
			MACs:         algorithms.MACs,
		},
		PublicKeyAuthAlgorithms: algorithms.PublicKeyAuths,
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if authorizedKeysMap[string(pubKey.Marshal())] {
				return &ssh.Permissions{
					// Record the public key used for authentication.
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}

	signer, err := ssh.NewSignerFromKey(cfg.SSHPrivateKey)
	if err != nil {
		return nil, help.WrapErr(err, "parsing SSH private key")
	}
	config.AddHostKey(signer)

	return config, nil
}
