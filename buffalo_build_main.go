package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	_ "github.com/ArnaudCalmettes/microsocial/actions"
	"github.com/gobuffalo/buffalo/runtime"
	"github.com/markbates/grift/grift"

	"github.com/ArnaudCalmettes/microsocial/models"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/pop"

	_ "github.com/ArnaudCalmettes/microsocial/grifts"
)

func init() {
	t, err := time.Parse(time.RFC3339, "2019-09-18T11:51:22+02:00")
	if err != nil {
		fmt.Println(err)
	}
	runtime.SetBuild(runtime.BuildInfo{
		Version: "94c2bef",
		Time:    t,
	})
}

func main() {
	args := os.Args
	var originalArgs []string
	for i, arg := range args {
		if arg == "--" {
			originalArgs = append([]string{args[0]}, args[i+1:]...)
			args = args[:i]
			break
		}
	}
	if len(args) == 1 {
		if len(originalArgs) != 0 {
			os.Args = originalArgs
		}
		originalMain()
		return
	}
	c := args[1]
	switch c {

	case "migrate":
		migrate()

	case "version":
		printVersion()
	case "task", "t", "tasks":
		if len(args) < 3 {
			log.Fatal("not enough arguments passed to task")
		}
		c := grift.NewContext(args[2])
		if len(args) > 2 {
			c.Args = args[3:]
		}
		err := grift.Run(args[2], c)
		if err != nil {
			log.Fatal(err)
		}
	default:
		if _, err := exec.LookPath("buffalo"); err != nil {
			if err != nil {
				log.Fatal(err)
			}
		}
		cmd := exec.Command("buffalo", args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func printVersion() {
	fmt.Printf("microsocial version %s\n", runtime.Build())
}

func migrate() {
	box, err := pop.NewMigrationBox(packr.New("app:migrations", "./migrations"), models.DB)
	if err != nil {
		log.Fatalf("Failed to unpack migrations: %s", err)
	}
	err = box.Up()
	if err != nil {
		log.Fatalf("Failed to run migrations: %s", err)
	}
}
