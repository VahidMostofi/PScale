package autoscaler

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/vahidmostofi/wise-auto-scaler/internal/aggregator"
	"golang.org/x/crypto/ssh"
)

type simpleAutoscaler struct {
	MonitorInterval time.Duration // in seconds
	ac              aggregator.RequestCounts
}

// Autoscale ...
func (sa *simpleAutoscaler) Autoscale(signals chan Signal, close chan bool, errs chan error) {
	fmt.Println("autoscaling with monitor interval of", sa.MonitorInterval, "seconds")
	go sa.monitorRequestCounts(close, errs)
}

func (sa *simpleAutoscaler) updateRequestRate(rate int) {
	if rate > 126 {
		fmt.Println("Autoscaler: scaling to 150")
		executeCommand("kubectl apply -f /home/vahid/workspace/dynamicworkload/configs/bookstore-nodejs/" + strconv.Itoa(175))
	} else if rate > 101 {
		fmt.Println("Autoscaler: scaling to 125")
		executeCommand("kubectl apply -f /home/vahid/workspace/dynamicworkload/configs/bookstore-nodejs/" + strconv.Itoa(150))
	} else if rate > 76 {
		fmt.Println("Autoscaler: scaling to 100")
		executeCommand("kubectl apply -f /home/vahid/workspace/dynamicworkload/configs/bookstore-nodejs/" + strconv.Itoa(125))
	} else if rate > 1 {
		fmt.Println("Autoscaler: scaling to 75")
		executeCommand("kubectl apply -f /home/vahid/workspace/dynamicworkload/configs/bookstore-nodejs/" + strconv.Itoa(100))
	}
}

func getHostKey(host string) (ssh.PublicKey, error) {
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				return nil, errors.New(fmt.Sprintf("error parsing %q: %v", fields[2], err))
			}
			break
		}
	}

	if hostKey == nil {
		return nil, errors.New(fmt.Sprintf("no hostkey for %s", host))
	}
	return hostKey, nil
}

func publicKeyFile(file string) ssh.Signer {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	fmt.Println("AA")
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	fmt.Println("AA")
	return key
}

func executeCommand(command string) string {
	var hostKey ssh.PublicKey

	hostKey, err := getHostKey("136.159.209.204")
	if err != nil {
		log.Fatal(err)
	}

	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	//
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	key, err := ioutil.ReadFile("/home/vahid/.ssh/id_rsa")
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "vahid",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	// Connect to the remote server and perform the SSH handshake.
	client, err := ssh.Dial("tcp", "server204:22", config)
	if err != nil {
		log.Fatalf("unable to connect: %v", err)
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()
	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	return b.String()
}

// monitor ...
func (sa *simpleAutoscaler) monitorRequestCounts(close chan bool, errs chan error) {
	ticker := time.NewTicker(sa.MonitorInterval)
	for {
		select {
		case <-ticker.C:
			end := time.Now().UnixNano() / 1e6
			start := end - sa.MonitorInterval.Nanoseconds()/1e6
			rc, err := sa.ac.GetRequestsCounts(start, end)
			if err != nil {
				errs <- err
				return
			}
			total := 0
			rNames, err := sa.ac.GetRequestsNames(start, end)
			if err != nil {
				errs <- err
				return
			}
			str := "("
			// convertor := int(sa.MonitorInterval.Seconds())
			for i, reqName := range rNames {
				total += int(rc[reqName] / int(sa.MonitorInterval.Seconds()))
				str += strconv.Itoa(rc[reqName] / int(sa.MonitorInterval.Seconds()))

				if i != len(rNames)-1 {
					str += ","
				}
			}
			str += ")"
			str += strconv.Itoa(total)
			fmt.Println(str)
			sa.updateRequestRate(total)
		case <-close:
			fmt.Println("signal to stop simple autoscaling")
			return
		}
	}
}

// GetNewSimpleAutoscaler ...
func GetNewSimpleAutoscaler() (Autoscaler, error) {
	s := &simpleAutoscaler{}
	s.MonitorInterval = viper.GetDuration("autoscale_interval") * time.Second
	_, ac, _, err := aggregator.GetAll()
	if err != nil {
		panic(err)
	}
	s.ac = ac
	return s, nil
}
