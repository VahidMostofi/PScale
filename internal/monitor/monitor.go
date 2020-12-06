package monitor

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
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

// RecordPodDetails ...
func RecordPodDetails(outputFile string, close chan bool) {
	ticker := time.NewTicker(viper.GetDuration("pod_count_interval") * time.Second)
	fmt.Println("recording CPU count every", viper.GetDuration("pod_count_interval"), "second")
	go func() {
		for {
			select {
			case <-ticker.C:
				s := executeCommand("kubectl get pods -o yaml")
				b := []byte(s)
				res, _ := parse(b)

				var sum float64
				for _, value := range res {
					sum += value
				}
				f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					panic(err)
				}
				if _, err = f.WriteString(fmt.Sprintf("%d,%f\n", time.Now().Unix(), sum)); err != nil {
					panic(err)
				}
				f.Close()
			case <-close:
				return
			}
		}
	}()
}

func parse(b []byte) (map[string]float64, error) {
	temp := make(map[string]interface{})
	res := map[string]float64{"auth": 0, "entry": 0, "books": 0}
	yaml.Unmarshal(b, temp)
	for _, pt := range temp["items"].([]interface{}) {
		container := pt.(map[interface{}]interface{})["spec"].(map[interface{}]interface{})["containers"].([]interface{})[0].(map[interface{}]interface{})
		imgName := container["image"].(string)
		service := ""
		if strings.Contains(imgName, "_auth") {
			service = "auth"
		} else if strings.Contains(imgName, "_books") {
			service = "books"
		} else if strings.Contains(imgName, "_entry") {
			service = "entry"
		}
		if service != "" {
			cpuStr := container["resources"].(map[interface{}]interface{})["limits"].(map[interface{}]interface{})["cpu"].(string)
			if strings.Contains(cpuStr, "m") {
				cpuStr = cpuStr[:len(cpuStr)-1]
				cpuStr = "0." + cpuStr
			}
			cpuF, err := strconv.ParseFloat(cpuStr, 64)
			if err != nil {
				panic(err)
			}
			res[service] += cpuF
		}
	}
	return res, nil
}

func getHostKey(host string) (ssh.PublicKey, error) {
	// file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	file, err := os.Open(filepath.Join("/home/vahid", ".ssh", "known_hosts"))

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
