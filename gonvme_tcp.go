package gonvme

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	// ChrootDirectory allows the nvme commands to be run within a chrooted path, helpful for containerized services
	ChrootDirectory = "chrootDirectory"

	// DefaultInitiatorNameFile is the default file which contains the initiator nqn
	DefaultInitiatorNameFile = "/etc/nvme/hostnqn"

	// NVMeCommand - nvme command
	NVMeCommand = "nvme"

	// NVMePort - port number
	NVMePort = "4420"
)

// NVMeTCP provides many nvme-specific functions
type NVMeTCP struct {
	NVMeType
}

// NewNVMeTCP - returns a new NVMeTCP client
func NewNVMeTCP(opts map[string]string) *NVMeTCP {
	nvme := NVMeTCP{
		NVMeType: NVMeType{
			mock:    false,
			options: opts,
		},
	}

	return &nvme
}

func (nvme *NVMeTCP) getChrootDirectory() string {
	s := nvme.options[ChrootDirectory]
	if s == "" {
		s = "/"
	}
	return s
}

func (nvme *NVMeTCP) buildNVMeCommand(cmd []string) []string {
	if nvme.getChrootDirectory() == "/" {
		return cmd
	}
	command := []string{"chroot", nvme.getChrootDirectory()}
	command = append(command, cmd...)
	return command
}

// DiscoverNVMeTCPTargets - runs nvme discovery and returns a list of targets.
func (nvme *NVMeTCP) DiscoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	return nvme.discoverNVMeTCPTargets(address, login)
}

func (nvme *NVMeTCP) discoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	// TODO: add injection check on address
	// nvme discovery is done via nvme cli
	// nvme discover -t tcp -a <NVMe interface IP> -s <port>
	exe := nvme.buildNVMeCommand([]string{NVMeCommand, "discover", "-t", "tcp", "-a", address, "-s", NVMePort})
	cmd := exec.Command(exe[0], exe[1:]...)

	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("\nError discovering %s: %v", address, err)
		return []NVMeTarget{}, err
	}

	targets := make([]NVMeTarget, 0)
	nvmeTarget := NVMeTarget{}
	entryCount := 0
	skipIteration := false

	for _, line := range strings.Split(string(out), "\n") {
		// Output should look like:

		// Discovery Log Number of Records 2, Generation counter 2
		// =====Discovery Log Entry 0======
		// trtype:  fc
		// adrfam:  fibre-channel
		// subtype: nvme subsystem
		// treq:    not specified
		// portid:  0
		// trsvcid: none
		// subnqn:  nqn.1111-11.com.dell:powerstore:00:a1a1a1a111a1111a111a
		// traddr:  nn-0x11aaa111a1111a11:aa-0x11aaa11111111a11
		//
		// =====Discovery Log Entry 1======
		// trtype:  tcp
		// adrfam:  ipv4
		// subtype: nvme subsystem
		// treq:    not specified
		// portid:  2304
		// trsvcid: 4420
		// subnqn:  nqn.1111-11.com.dell:powerstore:00:a1a1a1a111a1111a111a
		// traddr:  1.1.1.1
		// sectype: none

		tokens := strings.Fields(line)
		if len(tokens) < 2 {
			continue
		}
		key := tokens[0]
		value := strings.Join(tokens[1:], " ")
		switch key {

		case "=====Discovery":
			// add to array
			if entryCount != 0 && !skipIteration {
				targets = append(targets, nvmeTarget)
			}
			nvmeTarget = NVMeTarget{}
			skipIteration = false
			entryCount++
			continue

		case "trtype:":
			nvmeTarget.TargetType = value
			if value != NVMeTransportTypeTCP {
				skipIteration = true
			}
			break

		case "traddr:":
			nvmeTarget.Portal = value
			break

		case "subnqn:":
			nvmeTarget.TargetNqn = value
			break

		case "adrfam:":
			nvmeTarget.AdrFam = value
			break

		case "subtype:":
			nvmeTarget.SubType = value
			break

		case "treq:":
			nvmeTarget.Treq = value
			break

		case "portid:":
			nvmeTarget.PortID = value
			break

		case "trsvcid:":
			nvmeTarget.TrsvcID = value
			break

		case "sectype:":
			nvmeTarget.SecType = value
			break

		default:

		}
	}
	if !skipIteration && nvmeTarget.TargetNqn != "" {
		targets = append(targets, nvmeTarget)
	}

	// TODO: Add optional login
	// log into the target if asked
	/*if login {
		for _, t := range targets {
			gonvme.PerformLogin(t)
		}
	}*/

	return targets, nil
}

// GetInitiators returns a list of initiators on the local system.
func (nvme *NVMeTCP) GetInitiators(filename string) ([]string, error) {
	return nvme.getInitiators(filename)
}

func (nvme *NVMeTCP) getInitiators(filename string) ([]string, error) {

	// a slice of filename, which might exist and define the nvme initiators
	initiatorConfig := []string{}
	nqns := []string{}

	if filename == "" {
		// add default filename(s) here
		// /etc/nvme/hostnqn is the proper file for CentOS, RedHat, Sles, Ubuntu
		if nvme.getChrootDirectory() != "/" {
			initiatorConfig = append(initiatorConfig, nvme.getChrootDirectory()+"/"+DefaultInitiatorNameFile)
		} else {
			initiatorConfig = append(initiatorConfig, DefaultInitiatorNameFile)
		}
	} else {
		initiatorConfig = append(initiatorConfig, filename)
	}

	var err error
	// for each initiatior config file
	for _, init := range initiatorConfig {
		// make sure the file exists
		_, err = os.Stat(init)
		if err != nil {
			continue
		}

		// get the contents of the initiator config file
		// TODO: check if sys call is available for cat command
		cmd := exec.Command("cat", init)

		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error gathering initiator names: %v", err)
		}
		lines := strings.Split(string(out), "\n")

		for _, line := range lines {

			if line != "" {
				nqns = append(nqns, line)
			}
		}
	}

	if len(nqns) == 0 {
		return nqns, err
	}

	return nqns, nil
}

// NVMeConnect will attempt to connect into a given nvme target
func (nvme *NVMeTCP) NVMeConnect(target NVMeTarget) error {
	return nvme.nvmeConnect(target)
}

func (nvme *NVMeTCP) nvmeConnect(target NVMeTarget) error {
	// nvme connect is done via the nvme cli
	// nvme connect -t tcp -n <target NQN> -a <NVMe interface IP> -s 4420
	exe := nvme.buildNVMeCommand([]string{NVMeCommand, "connect", "-t", "tcp", "-n", target.TargetNqn, "-a", target.Portal, "-s", NVMePort})
	cmd := exec.Command(exe[0], exe[1:]...)

	_, err := cmd.Output()

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// nvme connect exited with an exit code != 0
			nvmeConnectResult := -1
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				nvmeConnectResult = status.ExitStatus()
			}
			if nvmeConnectResult == 114 {
				// session already exists
				// do not treat this as a failure
				fmt.Printf("\nnvme connection already exists to: %s", target.TargetNqn)
				err = nil
			} else {
				fmt.Printf("\nnvme connect failure: %v", err)
			}
		} else {
			fmt.Printf("\nError during nvme connect %s at %s: %v", target.TargetNqn, target.Portal, err)
		}

		if err != nil {
			fmt.Printf("\nError during nvme connect %s at %s: %v", target.TargetNqn, target.Portal, err)
			return err
		}
	} else {
		fmt.Printf("\nnvme connect successful: %s", target.TargetNqn)
	}

	return nil
}

// NVMeDisonnect will attempt to disconnect from a given nvme target
func (nvme *NVMeTCP) NVMeDisonnect(target NVMeTarget) error {
	return nvme.nvmeDisonnect(target)
}

func (nvme *NVMeTCP) nvmeDisonnect(target NVMeTarget) error {
	// nvme disconnect is done via the nvme cli
	// nvme disconnect -n <target NQN>
	exe := nvme.buildNVMeCommand([]string{NVMeCommand, "disconnect", "-n", target.TargetNqn})
	cmd := exec.Command(exe[0], exe[1:]...)

	_, err := cmd.Output()

	if err != nil {
		fmt.Printf("\nError logging %s at %s: %v", target.TargetNqn, target.Portal, err)
	} else {
		fmt.Printf("\nnvme disconnect successful: %s", target.TargetNqn)
	}

	return err
}