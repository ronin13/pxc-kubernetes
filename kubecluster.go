package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type nodeConfig struct {
	Bstrap    string
	Joiner    string
	NodeCount int
}

var (
	createCluster bool
	deleteCluster bool
	runNode       bool
	zone          string
	clusterName   string
	projectID     string
	nd            nodeConfig
	stopNode      string
)

const nodeTempl = `
{
"id": "node{{.NodeCount}}",
"kind": "Pod",
"apiVersion": "v1beta1",
"desiredState": {
  "manifest": {
    "version": "v1beta1",
    "id": "node{{.NodeCount}}",
    "containers": [{
      "name": "node{{.NodeCount}}",
      "image": "ronin/pxc:centos7-release",
      "ports": [{ "containerPort": 3306 }, {"containerPort": 4567 }, {"containerPort": 4568 } ],
      "command": ["/usr/sbin/mysqld",  "--basedir=/usr",  "--wsrep-node-name=node{{.NodeCount}}",   "--user=mysql", {{.Bstrap}}  "--skip-grant-tables", "--wsrep_cluster_address=gcomm://{{.Joiner}}", "--wsrep-sst-method=rsync"]
    }]
  }
},
"labels": { 
    "name": "pxc", "node": "node{{.NodeCount}}"
  }
}
`

func runWithMsg(cmd string, msg string) string {
	var rval []byte
	var err error
	log.Printf("Running %s", cmd)

	cmnd := exec.Command("sh", "-c", cmd)
	cmnd.Stderr = os.Stderr
	//cmnd.Stdin = os.Stdin
	if rval, err = cmnd.Output(); err != nil {
		if len(msg) > 0 {
			log.Panicf(fmt.Sprintf("Message %s, Error %s, Command %s", msg, err, cmd))
		}
	}
	return string(rval)
}

func getCount() int {
	var cnt int
	var err error

	str := strings.Replace(runWithMsg("kubectl get pods --no-headers=true  -l 'name=pxc' | wc -l", "gcloud invocation failed"), "\n", "", -1)
	if cnt, err = strconv.Atoi(str); err != nil {
		log.Panicf("Failed to get count due to %s", err)
	}
	cnt = cnt / 2
	log.Printf("%d nodes are up", cnt)
	return cnt
}

func parseFlags() {
	flag.BoolVar(&createCluster, "create", false, "Create cluster")
	flag.StringVar(&clusterName, "name", "pxc-cluster", "Name of cluster")
	flag.StringVar(&zone, "zone", "us-central1-a", "Cluster zone")
	flag.BoolVar(&deleteCluster, "delete", false, "Delete cluster")
	flag.BoolVar(&runNode, "start", false, "To start a node in the cluster")
	flag.StringVar(&stopNode, "stop", "", "Node to stop in cluster")
	flag.StringVar(&projectID, "project", "", "Project Id on GCE")
}

func cleanUp() {
	if err := recover(); err != nil {
		runWithMsg("kubectl stop -f cluster.json", "Failed to stop service")
		runWithMsg("kubectl stop pods -l 'name=pxc'", "Failed to stop pods")
		os.Exit(1)
	}
	log.Printf(runWithMsg("kubectl get pods -l 'name=pxc'", "Failed to get pods"))
}

func main() {
	var pipe io.WriteCloser
	var err error

	log.Println("Lets begin")

	log.Println("Make sure you are running >= 2015.05.12 of gcloud and compute")

	parseFlags()
	flag.Parse()

	if projectID == "" {
		log.Fatalf("ProjectId is required: --project")
	}

	if stopNode != "" {
		clusterCmd := fmt.Sprintf("kubectl stop pods %s", stopNode)
		log.Println(runWithMsg(clusterCmd, "Failed to stop node"))
		os.Exit(0)
	}

	if deleteCluster {
		log.Printf("Deleting cluster %s in zone %s", clusterName, zone)
		clusterCmd := fmt.Sprintf("gcloud alpha container clusters delete %s --zone %s --quiet", clusterName, zone)
		log.Println(runWithMsg(clusterCmd, "Failed to delete cluster"))
		os.Exit(0)
	}

	if createCluster {
		log.Printf("Creating cluster %s in zone %s", clusterName, zone)
		clusterCmd := fmt.Sprintf("gcloud alpha container clusters create %s --zone %s", clusterName, zone)
		log.Println(runWithMsg(clusterCmd, "Failed to create cluster"))
		runWithMsg(fmt.Sprintf("gcloud config set container/cluster %s", clusterName), "Failed to set cluster config")
		runWithMsg(fmt.Sprintf("kubectl config use-context gke_%s_%s_%s", projectID, zone, clusterName), "Failed to set kubectl config")
		runWithMsg(fmt.Sprintf("gcloud alpha container clusters describe %s", clusterName), "Failed to get cluster config")
		if !runNode {
			os.Exit(0)
		}
		time.Sleep(time.Second * 2)
	}

	check := runWithMsg("kubectl get services  -l 'type=cluster'", "gcloud invocation failed")

	if !strings.Contains(check, "cluster") {
		log.Println("Starting the cluster service")
		runWithMsg("kubectl create -f cluster.json", "Failed to spawn cluster service")
		runWithMsg("kubectl get services -l 'type=cluster'", "Failed to get services")
	}

	defer cleanUp()

	count := getCount()

	if count == 0 {
		nd = nodeConfig{"\"--wsrep-new-cluster\",", "", 0}
	} else {
		nd = nodeConfig{"", "cluster", count}
	}

	t := template.Must(template.New("nodeParser").Parse(nodeTempl))

	cmd := exec.Command("sh", "-c", "kubectl create -f -")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if pipe, err = cmd.StdinPipe(); err != nil {
		log.Panicf("Failed to get the stdin pipe for cmd %s due to %s", cmd.Path, err)
	}

	if err = cmd.Start(); err != nil {
		log.Panicf("%s failed to run from %s", cmd.Path, err)
	}

	log.Printf("Starting node%d with following configuration", count)
	if err = t.Execute(os.Stdout, nd); err != nil {
		log.Panicf("Template execution failed due to %s", err)
	}

	time.Sleep(time.Second * 2)

	if err = t.Execute(pipe, nd); err != nil {
		log.Panicf("Template execution failed due to %s", err)
	}

	pipe.Close()
	if err = cmd.Wait(); err != nil {
		log.Panicf("Template execution failed with %s", err)
	}
	time.Sleep(time.Second * 2)
	log.Printf("Successfully started node%d", count)
}
