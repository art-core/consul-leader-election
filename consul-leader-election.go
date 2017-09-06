package main

import (
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"os"
	"time"
	"flag"
	"log"
)

var key, keyValue, sessionName string
var healthChecks StringSliceFlag

func init() {
	healthChecks = StringSliceFlag{}
	flag.StringVar(&key, "key", "", "-key leader")
	flag.StringVar(&keyValue, "key-value", "", "-key-value value (Default: consul node name)")
	flag.StringVar(&sessionName, "session-name", "", "-session-name sessionName (Default: -key)")
	flag.Var(&healthChecks, "health-check", "-health-check service:serviceName (serfHealth is set by default)")
	flag.Parse()

	if key == "" {
		log.Fatal("argument -key is not set")
	}
}

func main() {
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		panic(err)
	}

	var sessionChecks []string
	sessionChecks = append(sessionChecks, "serfHealth")
	sessionChecks = append(sessionChecks, healthChecks...)

	var sessionEntryName string
	if sessionName != "" {
		sessionEntryName = sessionName
	} else {
		sessionEntryName = key
	}

	session := &consul.SessionEntry{
		Name: sessionEntryName,
		Checks: sessionChecks,
		LockDelay: (time.Duration(1) * time.Second),
	}

	sessionID, _, err := client.Session().Create(session, &consul.WriteOptions{})
	if err != nil {
		panic(err)
	}

	localNodeName, err := client.Agent().NodeName()
	if err != nil {
		panic(err)
	}

	var value string
	if keyValue != "" {
		value = keyValue
	} else {
		value = localNodeName
	}

	kvPair := &consul.KVPair{
		Key: key,
		Value: []byte(value),
		Session: sessionID,
	}

	success, _, err := client.KV().Acquire(kvPair, &consul.WriteOptions{})
	if err != nil {
		panic(err)
	}

	if success {
		fmt.Println("I'm the current leader.")
		os.Exit(0) // PASSING
	} else {
		client.Session().Destroy(sessionID, &consul.WriteOptions{})
	}

	kv, _, err := client.KV().Get(key, &consul.QueryOptions{})
	if err != nil {
		panic(err)
	}

	sessionInfo, _, err := client.Session().Info(kv.Session, &consul.QueryOptions{})
	if err != nil {
		panic(err)
	}

	if sessionInfo.Node == localNodeName {
		fmt.Println("I am the current leader.")
		os.Exit(0) // PASSING
	} else {
		fmt.Println(fmt.Sprintf("%s is the current leader.", sessionInfo.Node))
		os.Exit(2) // CRITICAL
	}

}
