package server

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/kelda/kelda/api"
	"github.com/kelda/kelda/api/client"
	"github.com/kelda/kelda/api/client/mocks"
	"github.com/kelda/kelda/api/pb"
	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/connection"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/minion/vault"
	vaultMocks "github.com/kelda/kelda/minion/vault/mocks"
	"github.com/stretchr/testify/assert"
)

func checkQuery(t *testing.T, s *server, table db.TableType, exp string) {
	reply, err := s.Query(context.Background(),
		&pb.DBQuery{Table: string(table)})

	assert.NoError(t, err)
	assert.Equal(t, exp, reply.TableContents, "Wrong query response")
}

func TestQueryErrors(t *testing.T) {
	// Invalid table type.
	s := server{}
	_, err := s.Query(context.Background(),
		&pb.DBQuery{Table: string(db.HostnameTable)})
	assert.EqualError(t, err, "unrecognized table: db.Hostname")
}

func TestQueryMachinesDaemon(t *testing.T) {
	t.Parallel()

	conn := db.New()
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		m := view.InsertMachine()
		m.Role = db.Master
		m.Provider = db.Amazon
		m.Size = "size"
		m.PublicIP = "8.8.8.8"
		m.PrivateIP = "9.9.9.9"
		m.Status = db.Connected
		view.Commit(m)

		return nil
	})

	exp := `[{"ID":1,"Role":"Master","Provider":"Amazon",` +
		`"Region":"","Size":"size","DiskSize":0,"SSHKeys":null,"FloatingIP":"",` +
		`"Preemptible":false,"CloudID":"","PublicIP":"8.8.8.8",` +
		`"PrivateIP":"9.9.9.9","Status":"connected","PublicKey":""}]`

	checkQuery(t, &server{conn: conn, runningOnDaemon: true}, db.MachineTable, exp)
}

func TestQueryContainersCluster(t *testing.T) {
	t.Parallel()

	conn := db.New()
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		c := view.InsertContainer()
		c.DockerID = "docker-id"
		c.Image = "image"
		c.Command = []string{"cmd", "arg"}
		view.Commit(c)

		return nil
	})

	exp := `[{"DockerID":"docker-id","Command":["cmd","arg"],` +
		`"Created":"0001-01-01T00:00:00Z","Image":"image"}]`

	checkQuery(t, &server{conn: conn, runningOnDaemon: false}, db.ContainerTable, exp)
}

func TestGetClusterContainers(t *testing.T) {
	newClient = func(host string, _ connection.Credentials) (client.Client, error) {
		switch host {
		case api.RemoteAddress("9.9.9.9"):
			mc := new(mocks.Client)
			mc.On("QueryContainers").Return([]db.Container{{
				BlueprintID: "onWorker",
				Image:       "shouldIgnore",
				DockerID:    "dockerID",
			}}, nil)
			mc.On("Close").Return(nil)
			return mc, nil
		default:
			t.Fatalf("Unexpected call to getClient with host %s", host)
		}
		panic("unreached")
	}

	mockLeaderClient := new(mocks.Client)
	mockLeaderClient.On("QueryContainers").Return([]db.Container{{
		BlueprintID: "notScheduled",
		Image:       "notScheduled",
	}, {
		BlueprintID: "onWorker",
		Image:       "onWorker",
	}}, nil)
	mockLeaderClient.On("Close").Return(nil)

	conn := db.New()
	conn.Txn(db.MachineTable).Run(func(view db.Database) error {
		m := view.InsertMachine()
		m.PublicIP = "9.9.9.9"
		m.Role = db.Worker
		m.Status = db.Connected
		view.Commit(m)

		return nil
	})

	s := server{conn: conn}
	actualContainers, err := s.getClusterContainers(mockLeaderClient)
	assert.NoError(t, err)

	expContainers := []db.Container{
		{BlueprintID: "notScheduled", Image: "notScheduled"},
		{BlueprintID: "onWorker", DockerID: "dockerID", Image: "onWorker"},
	}
	assert.Equal(t, expContainers, actualContainers)
}

func TestBadDeployment(t *testing.T) {
	conn := db.New()
	s := server{conn: conn, runningOnDaemon: true}

	badDeployment := `{`

	_, err := s.Deploy(context.Background(),
		&pb.DeployRequest{Deployment: badDeployment})

	assert.EqualError(t, err,
		"unable to parse blueprint: unexpected end of JSON input")
}
func TestInvalidImage(t *testing.T) {
	conn := db.New()
	s := &server{conn: conn, runningOnDaemon: true}
	testInvalidImage(t, s, "has:morethan:two:colons",
		"could not parse container image has:morethan:two:colons: "+
			"invalid reference format")
	testInvalidImage(t, s, "has-empty-tag:",
		"could not parse container image has-empty-tag:: "+
			"invalid reference format")
	testInvalidImage(t, s, "has-empty-tag::digest",
		"could not parse container image has-empty-tag::digest: "+
			"invalid reference format")
	testInvalidImage(t, s, "hasCapital",
		"could not parse container image hasCapital: "+
			"invalid reference format: repository name must be lowercase")
}

func testInvalidImage(t *testing.T, s *server, img, expErr string) {
	deployment := fmt.Sprintf(`
	{"Containers":[
		{"ID": "1",
                "Image": {"Name": "%s"},
                "Command":[
                        "sleep",
                        "10000"
                ],
                "Env": {}
	}]}`, img)

	_, err := s.Deploy(context.Background(),
		&pb.DeployRequest{Deployment: deployment})
	assert.EqualError(t, err, expErr)
}

func TestDeploy(t *testing.T) {
	conn := db.New()
	s := server{conn: conn, runningOnDaemon: true}

	createMachineDeployment := `
	{"Machines":[
		{"Provider":"Amazon",
		"Role":"Master",
		"Size":"m4.large"
	}, {"Provider":"Amazon",
		"Role":"Worker",
		"Size":"m4.large"
	}]}`

	_, err := s.Deploy(context.Background(),
		&pb.DeployRequest{Deployment: createMachineDeployment})

	assert.NoError(t, err)

	var bp db.Blueprint
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		bp, err = view.GetBlueprint()
		assert.NoError(t, err)
		return nil
	})

	exp, err := blueprint.FromJSON(createMachineDeployment)
	assert.NoError(t, err)
	assert.Equal(t, exp, bp.Blueprint)
}

func TestVagrantDeployment(t *testing.T) {
	conn := db.New()
	s := server{conn: conn, runningOnDaemon: true}

	vagrantDeployment := `
	{"Machines":[
		{"Provider":"Vagrant",
		"Role":"Master",
		"Size":"m4.large"
	}, {"Provider":"Vagrant",
		"Role":"Worker",
		"Size":"m4.large"
	}]}`
	vagrantErrMsg := "The Vagrant provider is still in development." +
		" The blueprint will continue to run, but" +
		" there may be some errors."

	_, err := s.Deploy(context.Background(),
		&pb.DeployRequest{Deployment: vagrantDeployment})

	assert.Error(t, err, vagrantErrMsg)

	var bp db.Blueprint
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		bp, err = view.GetBlueprint()
		assert.NoError(t, err)
		return nil
	})

	exp, err := blueprint.FromJSON(vagrantDeployment)
	assert.NoError(t, err)
	assert.Equal(t, exp, bp.Blueprint)
}

func TestUpdateLeaderContainerAttrs(t *testing.T) {
	t.Parallel()

	created := time.Now()

	lContainers := []db.Container{
		{
			BlueprintID: "1",
		},
	}

	wContainers := []db.Container{
		{
			BlueprintID: "1",
			Created:     created,
			Status:      "running",
		},
	}

	// Test update a matching container.
	expect := wContainers
	result := updateLeaderContainerAttrs(lContainers, wContainers)
	assert.Equal(t, expect, result)

	// Test container in leader, not in worker.
	newContainer := db.Container{
		BlueprintID: "2",
	}
	lContainers = append(lContainers, newContainer)
	expect = append(expect, newContainer)
	result = updateLeaderContainerAttrs(lContainers, wContainers)
	assert.Equal(t, expect, result)

	// Test if wContainers empty.
	lContainers = wContainers
	wContainers = []db.Container{}
	expect = lContainers
	result = updateLeaderContainerAttrs(lContainers, wContainers)
	assert.Equal(t, expect, result)

	// Test if wContainers and lContainers empty.
	lContainers = []db.Container{}
	expect = nil
	result = updateLeaderContainerAttrs(lContainers, wContainers)
	assert.Equal(t, expect, result)

	// Test a deployed Dockerfile.
	lContainers = []db.Container{{BlueprintID: "1", Image: "image"}}
	wContainers = []db.Container{
		{BlueprintID: "1", Image: "8.8.8.8/image", Created: created},
	}
	expect = []db.Container{{BlueprintID: "1", Image: "image", Created: created}}
	result = updateLeaderContainerAttrs(lContainers, wContainers)
	assert.Equal(t, expect, result)
}

func TestDaemonOnlyEndpoints(t *testing.T) {
	t.Parallel()

	s := server{}
	_, err := s.QueryMinionCounters(nil, nil)
	assert.EqualError(t, err, errDaemonOnlyRPC.Error())

	_, err = s.Deploy(nil, nil)
	assert.EqualError(t, err, errDaemonOnlyRPC.Error())
}

func TestQueryImagesCluster(t *testing.T) {
	t.Parallel()

	conn := db.New()
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		img := view.InsertImage()
		img.Name = "foo"
		view.Commit(img)

		return nil
	})

	exp := `[{"ID":1,"Name":"foo","Dockerfile":"","DockerID":"","Status":""}]`
	checkQuery(t, &server{conn: conn}, db.ImageTable, exp)
}

// The Daemon should get a connection to the leader of the cluster, and
// forward the secret association.
func TestSetSecretDaemon(t *testing.T) {
	secretName := "secretName"
	secretValue := "secretValue"

	mc := new(mocks.Client)
	mc.On("SetSecret", secretName, secretValue).Return(nil)
	mc.On("Close").Return(nil)
	newLeaderClient = func(_ []db.Machine, _ connection.Credentials) (
		client.Client, error) {
		return mc, nil
	}

	s := server{conn: db.New(), runningOnDaemon: true}
	_, err := s.SetSecret(nil, &pb.Secret{
		Name: secretName, Value: secretValue,
	})
	assert.NoError(t, err)
	mc.AssertExpectations(t)
}

// The minion should get a connection to Vault, and write the secret.
func TestSetSecretCluster(t *testing.T) {
	secretName := "secretName"
	secretValue := "secretValue"
	myIP := "1.2.3.4"

	conn := db.New()
	conn.Txn(db.MinionTable).Run(func(view db.Database) error {
		m := view.InsertMinion()
		m.Self = true
		m.PrivateIP = myIP
		view.Commit(m)
		return nil
	})

	mockClient := &vaultMocks.SecretStore{}
	newVaultClient = func(addr string) (vault.SecretStore, error) {
		assert.Equal(t, myIP, addr)
		return mockClient, nil
	}

	mockClient.On("Write", secretName, secretValue).Return(nil).Once()
	s := server{conn: conn}
	_, err := s.SetSecret(nil, &pb.Secret{
		Name: secretName, Value: secretValue,
	})
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSyncClusterInfo(t *testing.T) {
	expImages := []db.Image{
		{Name: "test1"},
		{Name: "test2"},
	}
	expImagesJSON, err := json.Marshal(expImages)
	assert.NoError(t, err)

	expLoadBalancers := []db.LoadBalancer{
		{Name: "test", IP: "ip", Hostnames: []string{"h1", "h2"}},
	}
	expLoadBalancersJSON, err := json.Marshal(expLoadBalancers)
	assert.NoError(t, err)

	mc := new(mocks.Client)
	mc.On("QueryImages").Return(expImages, nil)
	mc.On("QueryLoadBalancers").Return(expLoadBalancers, nil)
	mc.On("Close").Return(nil)

	// Test that a failure to retrieve one table doesn't affect the others.
	mc.On("QueryConnections").Return(nil, assert.AnError)

	// Don't test querying containers because the mocking required for querying
	// containers is more complicated, and is covered by TestGetClusterContainers.
	mc.On("QueryContainers").Return(nil, nil)

	newLeaderClient = func(_ []db.Machine, _ connection.Credentials) (
		client.Client, error) {
		return mc, nil
	}

	// There must be at least one connected machine in the database or else
	// the code won't attempt to connect to the cluster.
	conn := db.New()
	conn.Txn(db.MachineTable).Run(func(view db.Database) error {
		m := view.InsertMachine()
		m.Status = db.Connected
		view.Commit(m)
		return nil
	})

	apiServer := &server{
		conn:            conn,
		clusterInfo:     map[db.TableType]interface{}{},
		runningOnDaemon: true,
	}
	apiServer.syncClusterInfoOnce()

	checkQuery(t, apiServer, db.ImageTable, string(expImagesJSON))
	checkQuery(t, apiServer, db.LoadBalancerTable, string(expLoadBalancersJSON))
	checkQuery(t, apiServer, db.ConnectionTable, "null")
	checkQuery(t, apiServer, db.ContainerTable, "null")
}
