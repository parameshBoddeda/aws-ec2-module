package test

import (
    "time"
    "bytes"
    "io/ioutil"
    "testing"
    // "os/exec"
    "os"
    // "fmt"
    // "strings"

    // "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"golang.org/x/crypto/ssh"
)

func TestFileExists(t *testing.T) {
    t.Parallel()
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{

		TerraformDir: "../",
        MaxRetries:   3,
        TimeBetweenRetries: 5 * time.Second,
	})

	terraform.InitAndApply(t, terraformOptions)

	defer terraform.Destroy(t, terraformOptions)

	publicIp := terraform.Output(t, terraformOptions, "public_ip")
	

    // Create an SSH client to run commands on the EC2 instance
    sshUsername := "ec2-user"
    sshKeyPath := "../ssmKey.pem"
    permissions := 0600 // Example: Set the permissions to 0600 (Owner read/write)

	err := os.Chmod(sshKeyPath, os.FileMode(permissions))
	if err != nil {
		t.Errorf("Error setting file permissions: %v", err)
	}

    privateKeyBytes, err := ioutil.ReadFile(sshKeyPath)
    if err != nil {
        t.Errorf("Error reading private key file: %v", err)
    }
    privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
    if err != nil {
        t.Errorf("Error parsing private key: %v", err)
    }
	
    sshConfig := &ssh.ClientConfig{
        User: sshUsername,
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(privateKey),
        },
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }
    
    sshClient, err := ssh.Dial("tcp", publicIp+":22", sshConfig)
    if err != nil {
        t.Errorf("Error establishing SSH connection: %v", err)
    }
    if sshClient != nil {
        t.Logf("Making SSH Connection is for  %v success", sshClient)
    }
    
    defer sshClient.Close()

    command := "ls -l ~"
    session, err := sshClient.NewSession()
    if err != nil {
        t.Errorf("Error creating SSH session: %v", err)
    }
    defer session.Close()
    var stdout bytes.Buffer
    session.Stdout = &stdout
    if err := session.Run(command); err != nil {
        t.Errorf("Error running command: %v", err)
    }
    output := stdout.String()
    t.Logf("List of files:\n%s", output)
}
