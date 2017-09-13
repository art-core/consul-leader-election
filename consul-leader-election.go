package main

import (
	"flag"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"log"
	"os"
	"time"
)

var key, keyValue, sessionName, serviceName, leaderTag, notLeaderTag string
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
	flag.StringVar(&serviceName, "service-name", "", "-service-name serviceName")
	flag.StringVar(&leaderTag, "leader-tag", "", "-leader-tag master")
	flag.StringVar(&notLeaderTag, "not-leader-tag", "", "-not-leader-tag slave")
	flag.Parse()

	if key == "" {
		log.Fatal("argument -key is not set")
	}

	if sessionName != "" {
		sessionName = key
	}

	if leaderTag != "" && serviceName == "" {
		log.Fatal("argument -service-name is not set. (Required if you set leader-tag)")
	}

	if notLeaderTag != "" && serviceName == "" {
		log.Fatal("argument -service-name is not set. (Required if you set -not-leader-tag)")
	}

	if serviceName != "" && leaderTag == "" && notLeaderTag == "" {
		log.Fatalf("'-service-name %s' has no effect without setting -leader-tag or -not-leader-tag", serviceName)
	}
}

func main() {
	// create a new client
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		log.Println(err.Error())
		os.Exit(errorExitCode)
	}

	// get the local agents node name
	localNodeName, err := client.Agent().NodeName()
	if err != nil {
		log.Println(err.Error())
		os.Exit(errorExitCode)
	}

	// try to get the key value pair, for the given key
	kv, _, err := client.KV().Get(key, &consul.QueryOptions{})
	if err != nil {
		log.Println(err.Error())
		os.Exit(errorExitCode)
	}

	// if the key exists and has a session
	if kv != nil && kv.Session != "" {
		// get keys session info
		sessionInfo, _, err := client.Session().Info(kv.Session, &consul.QueryOptions{})
		if err != nil {
			log.Println(err.Error())
			os.Exit(errorExitCode)
		}

		// if the session belongs to the local agent, he is the leader
		if sessionInfo.Node == localNodeName {
			log.Println("I am the current leader.")
			leaderAction(client)
		} else {
			log.Println(fmt.Sprintf("%s is the current leader.", sessionInfo.Node))
			notLeaderAction(client)
		}
	}

	// define session
	var sessionChecks []string
	sessionChecks = append(sessionChecks, "serfHealth")
	sessionChecks = append(sessionChecks, healthChecks...)

	session := &consul.SessionEntry{
		Name:      sessionName,
		Checks:    sessionChecks,
		LockDelay: (time.Duration(1) * time.Second),
	}

	// create session
	sessionID, _, err := client.Session().Create(session, &consul.WriteOptions{})
	if err != nil {
		log.Println(err.Error())
		os.Exit(errorExitCode)
	}

	// set value of the key
	var value string
	if keyValue != "" {
		value = keyValue
	} else {
		value = localNodeName
	}

	// define key value pair
	kvPair := &consul.KVPair{
		Key:     key,
		Value:   []byte(value),
		Session: sessionID,
	}

	// try to acquire the key, with the created session
	success, _, err := client.KV().Acquire(kvPair, &consul.WriteOptions{})
	if err != nil {
		log.Println(err.Error())
		os.Exit(errorExitCode)
	}

	// if successful, the local agent is the leader
	if success {
		log.Println("I'm the current leader.")
		leaderAction(client)
	} else {
		client.Session().Destroy(sessionID, &consul.WriteOptions{})
		log.Println("Failed to acquire key.")
		notLeaderAction(client)
	}
}

// defines what to do if leader
func leaderAction(client *consul.Client) {
	if err := updateTag(client, serviceName, leaderTag); err != nil {
		log.Println(err.Error())
		os.Exit(errorExitCode)
	}
	os.Exit(leaderExitCode)
}

// defines what to do if not leader
func notLeaderAction(client *consul.Client) {
	if err := updateTag(client, serviceName, notLeaderTag); err != nil {
		log.Println(err.Error())
		os.Exit(errorExitCode)
	}
	os.Exit(notLeaderExitCode)
}

// updates the service with the given tag
func updateTag(client *consul.Client, serviceName, tag string) error {
	agent := client.Agent()
	services, err := agent.Services()
	if err != nil {
		return err
	}

	service, serviceExists := services[serviceName]

	if !serviceExists {
		return fmt.Errorf("Service '%s' doesn't exist.", serviceName)
	}

	if inSlice(tag, service.Tags) {
		return nil
	}

	var tags []string
	if tag != "" {
		tags = append(cleanupTagSlice(service.Tags), tag)
	} else {
		tags = cleanupTagSlice(service.Tags)
	}

	serviceRegistration := &consul.AgentServiceRegistration{
		ID:                service.ID,
		Name:              service.Service,
		Tags:              tags,
		Port:              service.Port,
		Address:           service.Address,
		EnableTagOverride: service.EnableTagOverride,
	}

	if err := agent.ServiceRegister(serviceRegistration); err != nil {
		return err
	} else {
		return nil
	}
}

// removes the "-leader-tag" and "-not-leader-tag" from the given slice
func cleanupTagSlice(slice []string) []string {
	var result []string
	for _, v := range slice {
		if v == leaderTag || v == notLeaderTag {
			continue
		}
		result = append(result, v)
	}

	return result
}

func inSlice(element string, slice []string) bool {
	for _, el := range slice {
		if el == element {
			return true
		}
	}

	return false
}
