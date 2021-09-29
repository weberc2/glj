package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/mitchellh/go-homedir"
)

type commit struct {
	Hash         string
	Author       object.Signature
	Committer    object.Signature
	PGPSignature string
	Message      string
	TreeHash     string
	ParentHashes []string
}

func newCommit(c *object.Commit) commit {
	parentHashes := make([]string, len(c.ParentHashes))
	for i := range c.ParentHashes {
		parentHashes[i] = c.ParentHashes[i].String()
	}
	return commit{
		Hash:         c.Hash.String(),
		Author:       c.Author,
		Committer:    c.Committer,
		PGPSignature: c.PGPSignature,
		Message:      c.Message,
		TreeHash:     c.TreeHash.String(),
		ParentHashes: parentHashes,
	}
}

func main() {
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Error loading home directory: %v", err)
	}

	privateKeyFile := filepath.Join(homeDir, ".ssh", "id_rsa")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Missing required flag REPO")
	}
	url := args[0]

	// Load the public keys
	publicKeys, err := ssh.NewPublicKeysFromFile(
		"git",
		privateKeyFile,
		"",
	)
	if err != nil {
		log.Fatalf("Retrieving public keys from private key file: %v", err)
	}

	// Clone the repo into memory
	r, err := git.Clone(
		memory.NewStorage(),
		nil,
		&git.CloneOptions{
			URL:      url,
			Auth:     publicKeys,
			Progress: os.Stderr,
		},
	)
	if err != nil {
		log.Fatalf("Cloning repo: %v", err)
	}

	// Fetch the commit log
	commits, err := r.Log(&git.LogOptions{})
	if err != nil {
		log.Fatalf("Fetching log: %v", err)
	}

	if err := commits.ForEach(func(c *object.Commit) error {
		data, err := json.Marshal(newCommit(c))
		if err != nil {
			log.Fatalf("Serializing commit: %# v", c)
		}
		_, err = os.Stdout.Write(data)
		return err
	}); err != nil {
		log.Fatal(err)
	}
}
