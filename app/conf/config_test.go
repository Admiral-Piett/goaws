package conf

import (
    "github.com/Admiral-Piett/goaws/app"
    "github.com/Admiral-Piett/goaws/app/fixtures"
    "github.com/Admiral-Piett/goaws/app/utils"
    "github.com/mitchellh/copystructure"
    "github.com/stretchr/testify/assert"
    "os"
    "reflect"
    "sync"
    "testing"
)

func TestConfig_NewConfigLoader(t *testing.T) {
    c := NewConfigLoader("", "Local")

    assert.Contains(t, c.configFilename, "app/conf/goaws.yaml")
    assert.Equal(t, "Local", c.envTitle)
    assert.NotEqual(t, app.Environment{}, c.env)
    assert.Equal(t, []string{"4100"}, c.Ports)
    assert.Equal(t, reflect.ValueOf(createSQSQueues), reflect.ValueOf(c.createQueues))
    assert.Equal(t, reflect.ValueOf(createSNSTopics), reflect.ValueOf(c.createTopics))
}

func TestConfig_NewConfigLoader_usesSpecifiedConfig(t *testing.T) {
    c := NewConfigLoader("./mock-data/mock-config.yaml", "Local")

    assert.Contains(t, c.configFilename, "mock-data/mock-config.yaml")
    assert.Equal(t, "Local", c.envTitle)
    assert.NotEqual(t, app.Environment{}, c.env)
    assert.Equal(t, []string{"4200"}, c.Ports)
    assert.Equal(t, reflect.ValueOf(createSQSQueues), reflect.ValueOf(c.createQueues))
    assert.Equal(t, reflect.ValueOf(createSNSTopics), reflect.ValueOf(c.createTopics))
}

func TestConfig_FindDefaultConfig_success(t *testing.T) {
    c := &ConfigLoader{}
    c.findDefaultConfig()

    assert.Contains(t, c.configFilename, "app/conf/goaws.yaml")
}

func TestConfig_MissingConfigSetsDefaults(t *testing.T) {
    defer func() {
        app.CurrentEnvironment = app.Environment{}
    }()

    qCallCount := 0
    qs := func(queues []app.EnvQueue) error {
        qCallCount += 1
        return nil
    }
    tCallCount := 0
    ts := func(queues []app.EnvTopic) error {
        tCallCount += 1
        return nil
    }
    c := &ConfigLoader{
        configFilename: "./mock-data/mock-config.yaml",
        createQueues:   qs,
        createTopics:   ts,
    }

    c.loadYamlConfig("garbage")
    assert.Equal(t, []string{"4100"}, c.Ports)
    assert.Equal(t, fixtures.DEFAULT_ENVIRONMENT, c.env)
    assert.Equal(t, fixtures.DEFAULT_ENVIRONMENT, app.CurrentEnvironment)

    assert.Equal(t, 1, qCallCount)
    assert.Equal(t, 1, tCallCount)
}

func TestConfig_NoQueuesOrTopics(t *testing.T) {
    defer func() {
        app.CurrentEnvironment = app.Environment{}
    }()

    qCallCount := 0
    qs := func(queues []app.EnvQueue) error {
        qCallCount += 1
        return nil
    }
    tCallCount := 0
    ts := func(queues []app.EnvTopic) error {
        tCallCount += 1
        return nil
    }
    c := &ConfigLoader{
        configFilename: "./mock-data/mock-config.yaml",
        createQueues:   qs,
        createTopics:   ts,
    }

    c.loadYamlConfig("NoQueuesOrTopics")
    assert.Equal(t, []string{"4100"}, c.Ports)
    assert.Equal(t, fixtures.NO_QUEUES_NO_TOPICS_ENVIRONEMENT, c.env)
    assert.Equal(t, fixtures.NO_QUEUES_NO_TOPICS_ENVIRONEMENT, app.CurrentEnvironment)

    assert.Equal(t, 1, qCallCount)
    assert.Equal(t, 1, tCallCount)
}

func TestConfig_CreateQueuesTopicsAndSubscriptions(t *testing.T) {
    defer func() {
        app.CurrentEnvironment = app.Environment{}
    }()

    qCallCount := 0
    qs := func(queues []app.EnvQueue) error {
        qCallCount += 1
        return nil
    }
    tCallCount := 0
    ts := func(queues []app.EnvTopic) error {
        tCallCount += 1
        return nil
    }
    c := &ConfigLoader{
        configFilename: "./mock-data/mock-config.yaml",
        createQueues:   qs,
        createTopics:   ts,
    }

    c.loadYamlConfig("Local")
    assert.Equal(t, []string{"4200"}, c.Ports)
    assert.Equal(t, fixtures.LOCAL_ENVIRONMENT, c.env)
    assert.Equal(t, fixtures.LOCAL_ENVIRONMENT, app.CurrentEnvironment)

    assert.Equal(t, 1, qCallCount)
    assert.Equal(t, 1, tCallCount)
}

func TestConfig_HotReloadConfigFile(t *testing.T) {
    defer func() {
        app.CurrentEnvironment = app.Environment{}
    }()

    wg1 := sync.WaitGroup{}
    wg1.Add(2)

    qCallCount := 0
    qs := func(queues []app.EnvQueue) error {
        qCallCount += 1
        return nil
    }
    tCallCount := 0
    ts := func(queues []app.EnvTopic) error {
        tCallCount += 1
        wg1.Done()
        return nil
    }
    c := &ConfigLoader{
        configFilename: "./mock-data/mock-hotreload-config.yaml",
        createQueues:   qs,
        createTopics:   ts,
    }

    // Overwrite this file to make sure it's ready for the test
    configBackup, _ := os.ReadFile("./mock-data/mock-config.yaml")
    err := os.WriteFile(c.configFilename, configBackup, 0644)
    if err != nil {
        t.Errorf("Pre-write to %s failed", c.configFilename)
    }

    c.loadYamlConfig("Local")
    assert.Equal(t, []string{"4200"}, c.Ports)
    assert.Equal(t, fixtures.LOCAL_ENVIRONMENT, c.env)
    assert.Equal(t, fixtures.LOCAL_ENVIRONMENT, app.CurrentEnvironment)

    assert.Equal(t, 1, qCallCount)
    assert.Equal(t, 1, tCallCount)

    wg2 := &sync.WaitGroup{}
    wg2.Add(1)
    c.StartWatcher(wg2)
    wg2.Wait()

    configOverwrite := []byte("Local:\n  Host: not-localhost\n")
    err = os.WriteFile(c.configFilename, configOverwrite, 0644)
    if err != nil {
        t.Errorf("Overwrite to %s failed", c.configFilename)
    }

    wg1.Wait()

    assert.Equal(t, 2, qCallCount)
    assert.Equal(t, 2, tCallCount)

    err = os.WriteFile(c.configFilename, configBackup, 0644)
    if err != nil {
        t.Errorf("Reset write to %s failed", c.configFilename)
    }
}

func TestConfig_createSNSTopics(t *testing.T) {
    utils.ResetAppTopics()
    utils.ResetAppQueues()
    defer func() {
        utils.ResetAppTopics()
        utils.ResetAppQueues()
    }()

    err := createSNSTopics(fixtures.LOCAL_ENVIRONMENT.Topics)

    assert.Nil(t, err)

    utils.AssertTopicsMatchFixture(t, fixtures.LOCAL_APP_TOPICS)
    // These are created since they are the target of subscriptions
    utils.AssertQueuesMatchFixture(t, fixtures.LOCAL_APP_QUEUES)
}

func TestConfig_createSNSTopics_invalidFilterPolicyReturnsError(t *testing.T) {
    dupe, err := copystructure.Copy(fixtures.LOCAL_ENVIRONMENT.Topics)
    if err != nil {
        t.Error(err.Error())
    }
    topics, _ := dupe.([]app.EnvTopic)
    topics[0].Subscriptions[0].FilterPolicy = "{garbage]"

    err = createSNSTopics(topics)

    assert.Error(t, err)
}

//func TestConfig_createSqsQueues(t *testing.T){
//    expectedQueues := []string{
//        "local-queue1",
//        "local-queue2",
//        "local-queue3",
//        "local-queue4",
//        "local-queue5",
//        "local-queue3-dlq",
//    }
//
//    resetAppQueues()
//    defer func() {
//        resetAppQueues()
//        SQS_QUEUE_LOADER = createSQSQueues
//    }()
//
//    // Silo these requests for this method only, so we know exactly what we're doing
//    yamlLoaderSqsQueuesLoaderCallCount := 0
//    SQS_QUEUE_LOADER = func(env string) error {
//        yamlLoaderSqsQueuesLoaderCallCount += 1
//        return nil
//    }
//
//    env := "Local"
//    port := LoadYamlConfig("./mock-data/mock-config.yaml", env)
//
//    err := createSQSQueues(env)
//
//    assert.Nil(t, err)
//    assert.Equal(t, "4100", port[0])
//    assert.Equal(t, 1, yamlLoaderSqsQueuesLoaderCallCount)
//
//    queues := app.SyncQueues.Queues
//    assert.Equal(t, 6, len(queues))
//    for _, expectedName := range expectedQueues {
//        _, ok := queues[expectedName]
//        assert.True(t, ok)
//    }
//}
//
// TODO - @Admiral-Piett - Flesh out createSQSQueues tests
// TODO - @Admiral-Piett - Flesh out createHttpSubscription tests
// TODO - @Admiral-Piett - Flesh out createSqsSubscription tests
// TODO - @Admiral-Piett - Flesh out setQueueRedrivePolicy tests
