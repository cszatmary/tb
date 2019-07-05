package deps

import (
	"os"
	"runtime"

	"github.com/TouchBistro/tb/util"
	log "github.com/sirupsen/logrus"
)

// Dependency is an os dependency needed to run core-devtools
type Dependency struct {
	Name          string
	InstallCmd    []string
	BeforeInstall func() error
	AfterInstall  func() error
}

// Dependency names because magic strings suck
const (
	// XcodeSelect = "xcode-select"
	Brew       = "brew"
	Pgcli      = "pgcli"
	Jq         = "jq"
	Aws        = "aws"
	Lazydocker = "lazydocker"
	// Nvm = "nvm"
	Node   = "node"
	Yarn   = "yarn"
	Docker = "docker"
)

var deps = map[string]Dependency{
	// ROT IN HELL STEVE
	// XcodeSelect: Dependency{
	// 	Name:       "xcode-select -p",
	// 	InstallCmd: []string{"xcode-select", "--install"},
	// },
	Brew: Dependency{
		Name:       "brew",
		InstallCmd: []string{"/usr/bin/ruby", "-e", "\"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)\""},
	},
	Pgcli: Dependency{
		Name: "pgcli",
		BeforeInstall: func() error {
			_, err := util.Exec("brew", "tap", "dbcli/tap")
			return err
		},
		InstallCmd: []string{"brew", "install", "pgcli"},
	},
	Jq: Dependency{
		Name:       "jq",
		InstallCmd: []string{"brew", "install", "jq"},
	},
	Aws: Dependency{
		Name:       "aws",
		InstallCmd: []string{"brew", "install", "awscli"},
	},
	Lazydocker: Dependency{
		Name: "lazydocker",
		BeforeInstall: func() error {
			_, err := util.Exec("brew", "tap", "jesseduffield/lazydocker")
			return err
		},
		InstallCmd: []string{"brew", "install", "lazydocker"},
	},

	// Nvm: Dependency{
	// 	Name:       "nvm",
	// 	InstallCmd: []string{"brew", "install", "nvm"},
	// 	AfterInstall: func() error {
	// 		home := os.Getenv("HOME") // TODO: Make portable for uzi?
	// 		dirPath := fmt.Sprintf("%s/.nvm", home)
	// 		if !util.FileOrDirExists(dirPath) {
	// 			err := os.Mkdir(dirPath, os.ModeDir)
	// 			return err
	// 		}
	// 		for _, rcFile := range []string{".zshrc", ".bashrc"} {
	// 			rcPath := fmt.Sprintf("%s/%s", home, rcFile)
	// 			fmt.Printf("...adding nvm export to %s\n", rcPath)
	// 			err := util.AppendLineToFile(rcPath, "export NVM_DIR=\"$HOME/.nvm\"")
	// 			err = util.AppendLineToFile(rcPath, ". \"/usr/local/opt/nvm/nvm.sh\"")
	// 			if err != nil {
	// 				return err
	// 			}
	// 		}
	// 		return nil
	// 	},
	// },

	// TODO: Check that `which node` resolves to something like /Users/<user>/.nvm/version/node/<version>/bin/node
	Node: Dependency{
		Name:       "node",
		InstallCmd: []string{"nvm", "install", "stable"},
	},
	Yarn: Dependency{
		Name:       "yarn",
		InstallCmd: []string{"brew", "install", "yarn"},
	},
	Docker: Dependency{
		Name: "docker",
		BeforeInstall: func() error {
			_, err := util.Exec("brew", "tap", "caskroom/versions")
			return err
		},
		InstallCmd: []string{"brew", "cask", "install", "docker"},
	},
}

func Resolve(depNames ...string) error {
	log.Println("checking dependencies...")

	if runtime.GOOS != "darwin" {
		log.Println("tb currently supports Darwin (MacOS) only for installing dependencies.")
		log.Println("if you want to support other OSes, please make a pull request or tell Dev Acceleration.")
		os.Exit(1)
	}

	for _, depName := range depNames {
		dep, ok := deps[depName]

		if !ok {
			log.Fatalf("%s is not a valid dependency\n", depName)
		}

		if util.IsCommandAvailable(dep.Name) {
			log.Printf("%s was found.\n", dep.Name)
			continue
		} else {
			log.Printf("%s was NOT found.\n", dep.Name)
		}

		log.Printf("installing %s.\n", dep.Name)

		if dep.BeforeInstall != nil {
			err := dep.BeforeInstall()
			if err != nil {
				return err
			}
		}

		installCmd := dep.InstallCmd[0]
		installArgs := dep.InstallCmd[1:]

		_, err := util.Exec(installCmd, installArgs...)
		if err != nil {
			return err
		}

		if dep.AfterInstall != nil {
			err := dep.AfterInstall()
			if err != nil {
				return err
			}
		}

		log.Printf("finished installing %s.\n", dep.Name)
	}
	return nil
}
