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
var leaderExitCode, notLeaderExitCode, errorExitCode int
var healthChecks StringSliceFlag

func init() {
	healthChecks = StringSliceFlag{}
	flag.StringVar(&key, "key", "", "-key leader")
	flag.StringVar(&keyValue, "key-value", "", "-key-value value (Default: consul node name)")
	flag.StringVar(&sessionName, "session-name", "", "-session-name sessionName (Default: -key)")
	flag.IntVar(&leaderExitCode, "leader-exit-code", 0, "-leader-exit-code 0")
	flag.IntVar(&notLeaderExitCode, "not-leader-exit-code", 1, "-not-leader-exit-code 1")
	flag.IntVar(&errorExitCode, "error-exit-code", 2, "-error-exit-code 2")
	flag.Var(&healthChecks, "health-check", "-health-check service:serviceName (serfHealth is set by default)")
	flag.Parse()

	if key == "" {
		log.Fatal("argument -key is not set")
	}

	if sessionName != "" {
		sessionName = key
	}
}

func main() {
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(errorExitCode)
	}

	var sessionChecks []string
	sessionChecks = append(sessionChecks, "serfHealth")
	sessionChecks = append(sessionChecks, healthChecks...)

	session := &consul.SessionEntry{
		Name: sessionName,
		Checks: sessionChecks,
		LockDelay: (time.Duration(1) * time.Second),
	}

	sessionID, _, err := client.Session().Create(session, &consul.WriteOptions{})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(errorExitCode)
	}

	localNodeName, err := client.Agent().NodeName()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(errorExitCode)
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
		fmt.Println(err.Error())
		os.Exit(errorExitCode)
	}

	if success {
		fmt.Println("I'm the current leader.")
		os.Exit(leaderExitCode)
	} else {
		client.Session().Destroy(sessionID, &consul.WriteOptions{})
	}

	kv, _, err := client.KV().Get(key, &consul.QueryOptions{})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(errorExitCode)
	}

	sessionInfo, _, err := client.Session().Info(kv.Session, &consul.QueryOptions{})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(errorExitCode)
	}

	if sessionInfo.Node == localNodeName {
		fmt.Println("I am the current leader.")
		os.Exit(leaderExitCode)
	} else {
		fmt.Println(fmt.Sprintf("%s is the current leader.", sessionInfo.Node))
		os.Exit(notLeaderExitCode)
	}

}
