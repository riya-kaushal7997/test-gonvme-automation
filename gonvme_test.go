/*
 *
 * Copyright © 2022-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *      http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package gonvme

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

type testData struct {
	TCPPortal     string
	FCPortal      string
	Target        string
	FCHostAddress string
}

var (
	tcpTestPortal string
	fcTestPortal  string
	testTarget    string
	hostAddress   string
)

func reset() {
	testValuesFile, err := os.ReadFile("testdata/unittest_values.json")
	if err != nil {
		log.Infof("Error Reading the file: %s ", err)
	}
	var testValues testData
	err = json.Unmarshal(testValuesFile, &testValues)
	if err != nil {
		log.Infof("Error during unmarshal: %s", err)
	}
	tcpTestPortal = testValues.TCPPortal
	fcTestPortal = testValues.FCPortal
	testTarget = testValues.Target
	hostAddress = testValues.FCHostAddress

	GONVMEMock.InduceDiscoveryError = false
	GONVMEMock.InduceInitiatorError = false
	GONVMEMock.InduceTCPLoginError = false
	GONVMEMock.InduceFCLoginError = false
	GONVMEMock.InduceLogoutError = false
	GONVMEMock.InduceGetSessionsError = false
	GONVMEMock.InducedNVMeDeviceAndNamespaceError = false
	GONVMEMock.InducedNVMeNamespaceIDError = false
	GONVMEMock.InducedNVMeDeviceDataError = false
}

func TestPolymorphichCapability(t *testing.T) {
	reset()
	var c NVMEinterface
	// start off with a real implementation
	c = NewNVMe(map[string]string{})
	if c.isMock() {
		// this should not be a mock implementation
		t.Error("Expected a real implementation but got a mock one")
		return
	}
	// switch it to mock
	c = NewMockNVMe(map[string]string{})
	if !c.isMock() {
		// this should not be a real implementation
		t.Error("Expected a mock implementation but got a real one")
		return
	}
	// switch back to a real implementation
	c = NewNVMe(map[string]string{})
	if c.isMock() {
		// this should not be a mock implementation
		t.Error("Expected a real implementation but got a mock one")
		return
	}
}

func TestDiscoverNVMeTCPTargets(t *testing.T) {
	reset()
	c := NewNVMe(map[string]string{})
	_, err := c.DiscoverNVMeTCPTargets(tcpTestPortal, false)
	if err == nil {
		t.Error(err.Error())
	}
}

func TestDiscoverNVMeFCTargets(t *testing.T) {
	reset()
	c := NewNVMe(map[string]string{})
	_, err := c.DiscoverNVMeFCTargets(fcTestPortal, false)
	FCHostsInfo, err := c.getFCHostInfo()
	if err == nil && len(FCHostsInfo) != 0 {
		_, err := c.DiscoverNVMeFCTargets(fcTestPortal, false)
		if err == nil {
			t.Error(err.Error())
		}
	}
}

func TestNVMeTCPLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     tcpTestPortal,
		TargetNqn:  testTarget,
		TrType:     "tcp",
		AdrFam:     "ipv4",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "tcp",
	}
	err := c.NVMeTCPConnect(tgt, false)
	if err == nil {
		t.Error(err.Error())
		return
	}
	nvmeSessions, _ := c.GetSessions()
	if len(nvmeSessions) != 0 {
		err = c.NVMeDisconnect(tgt)
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
}

func TestNVMeFCLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     fcTestPortal,
		TargetNqn:  testTarget,
		TrType:     "fc",
		AdrFam:     "fibre-channel",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "fc",
		HostAdr:    hostAddress,
	}
	err := c.NVMeFCConnect(tgt, false)
	if err == nil {
		t.Error(err.Error())
		return
	}
	nvmeSessions, _ := c.GetSessions()
	if len(nvmeSessions) != 0 {
		err = c.NVMeDisconnect(tgt)
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
}

func TestLoginLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     tcpTestPortal,
		TargetNqn:  testTarget,
		TrType:     "tcp",
		AdrFam:     "ipv4",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "tcp",
	}
	err := c.NVMeTCPConnect(tgt, false)
	if err == nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeFCConnect(tgt, false)
	if err == nil {
		t.Error(err.Error())
		return
	}
	nvmeSessions, _ := c.GetSessions()
	if len(nvmeSessions) != 0 {
		err = c.NVMeDisconnect(tgt)
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
}

func TestLogoutLogoutTargets(t *testing.T) {
	reset()
	c := NewNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     tcpTestPortal,
		TargetNqn:  testTarget,
		TrType:     "tcp",
		AdrFam:     "fibre-channel",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "tcp",
	}
	// log out of the target, just in case we are logged in already
	_ = c.NVMeTCPConnect(tgt, false)
	nvmeSessions, _ := c.GetSessions()
	if len(nvmeSessions) != 0 {
		err := c.NVMeDisconnect(tgt)
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
}

func TestGetInitiators(t *testing.T) {
	reset()
	testdata := []struct {
		filename string
		count    int
	}{
		{"testdata/initiatorname.nvme", 1},
		{"testdata/multiple_nqn.nvme", 2},
		{"testdata/no_nqn.nvme", 0},
		{"testdata/valid.nvme", 1},
	}

	c := NewNVMe(map[string]string{})
	for _, tt := range testdata {
		initiators, err := c.GetInitiators(tt.filename)
		if err != nil {
			t.Errorf("Error getting %d initiators from %s: %s", tt.count, tt.filename, err.Error())
		}
		if len(initiators) != tt.count {
			t.Errorf("Expected %d initiators in %s, but got %d", tt.count, tt.filename, len(initiators))
		}
	}
}

func TestBuildNVMECommand(t *testing.T) {
	reset()
	opts := map[string]string{}
	initial := []string{"/bin/ls"}
	opts[ChrootDirectory] = "/test"
	c := NewNVMe(opts)
	command := c.buildNVMeCommand(initial)
	// the length of the resulting command should the length of the initial command +2
	if len(command) != (len(initial) + 2) {
		t.Errorf("Expected to %d items in the command slice but received %v", len(initial)+2, command)
	}
	if command[0] != "chroot" {
		t.Error("Expected the command to be run with chroot")
	}
	if command[1] != opts[ChrootDirectory] {
		t.Errorf("Expected the command to chroot to %s but got %s", opts[ChrootDirectory], command[1])
	}
}

func TestListNVMeDeviceAndNamespace(t *testing.T) {
	reset()
	c := NewNVMe(map[string]string{})
	_, err := c.ListNVMeDeviceAndNamespace()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetNVMeDeviceData(t *testing.T) {
	reset()
	c := NewNVMe(map[string]string{})
	devicesAndNamespaces, _ := c.ListNVMeDeviceAndNamespace()

	if len(devicesAndNamespaces) > 0 {
		for _, device := range devicesAndNamespaces {
			DevicePath := device.DevicePath
			_, _, err := c.GetNVMeDeviceData(DevicePath)
			if err != nil {
				t.Error(err.Error())
			}
		}
	}
}

func TestListNVMeNamespaceID(t *testing.T) {
	reset()
	c := NewNVMe(map[string]string{})
	devicesAndNamespaces, _ := c.ListNVMeDeviceAndNamespace()

	if len(devicesAndNamespaces) > 0 {
		_, err := c.ListNVMeNamespaceID(devicesAndNamespaces)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

func TestGetSessions(t *testing.T) {
	reset()
	c := NewNVMe(map[string]string{})
	_, err := c.GetSessions()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestMockDiscoverNVMETCPTargets(t *testing.T) {
	reset()
	var c NVMEinterface
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfTCPTargets] = fmt.Sprintf("%d", expected)
	c = NewMockNVMe(opts)
	// c = mock
	targets, err := c.DiscoverNVMeTCPTargets("1.1.1.1", true)
	if err != nil {
		t.Error(err.Error())
	}
	if len(targets) != expected {
		t.Errorf("Expected to find %d targets, but got back %v", expected, targets)
	}
}

func TestMockDiscoverNVMEFCTargets(t *testing.T) {
	reset()
	var c NVMEinterface
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfFCTargets] = fmt.Sprintf("%d", expected)
	c = NewMockNVMe(opts)
	// c = mock
	targets, err := c.DiscoverNVMeFCTargets("nn-0x11aaa111111a1a1a:pn-0x11aaa111111a1a1a", true)
	if err != nil {
		t.Error(err.Error())
	}
	if len(targets) != expected {
		t.Errorf("Expected to find %d targets, but got back %v", expected, targets)
	}
}

func TestMockDiscoverNVMeTCPTargetsError(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfTCPTargets] = fmt.Sprintf("%d", expected)
	c := NewMockNVMe(opts)
	GONVMEMock.InduceDiscoveryError = true
	targets, err := c.DiscoverNVMeTCPTargets("1.1.1.1", false)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
	if len(targets) != 0 {
		t.Errorf("Expected to receive 0 targets when inducing an error. Received %v", targets)
		return
	}
}

func TestMockDiscoverNVMeFCTargetsError(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfFCTargets] = fmt.Sprintf("%d", expected)
	c := NewMockNVMe(opts)
	GONVMEMock.InduceDiscoveryError = true
	targets, err := c.DiscoverNVMeFCTargets("nn-0x11aaa111111a1a1a:pn-0x11aaa111111a1a1a", false)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
	if len(targets) != 0 {
		t.Errorf("Expected to receive 0 targets when inducing an error. Received %v", targets)
		return
	}
}

func TestMockGetInitiators(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 3
	opts[MockNumberOfInitiators] = fmt.Sprintf("%d", expected)
	c := NewMockNVMe(opts)
	initiators, err := c.GetInitiators("")
	if err != nil {
		t.Error(err.Error())
	}
	if len(initiators) != expected {
		t.Errorf("Expected to find %d initiators, but got back %v", expected, initiators)
	}
}

func TestMockGetInitiatorsError(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 3
	opts[MockNumberOfInitiators] = fmt.Sprintf("%d", expected)
	c := NewMockNVMe(opts)
	GONVMEMock.InduceInitiatorError = true
	initiators, err := c.GetInitiators("")
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
	if len(initiators) != 0 {
		t.Errorf("Expected to receive 0 initiators when inducing an error. Received %v", initiators)
		return
	}
}

func TestMockNVMeTCPLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     tcpTestPortal,
		TargetNqn:  testTarget,
		TrType:     "tcp",
		AdrFam:     "ipv4",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "tcp",
	}
	err := c.NVMeTCPConnect(tgt, false)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeDisconnect(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMockNVMeFCLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     fcTestPortal,
		TargetNqn:  testTarget,
		TrType:     "fc",
		AdrFam:     "fibre-channel",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "fc",
		HostAdr:    hostAddress,
	}
	err := c.NVMeFCConnect(tgt, false)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeDisconnect(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMockLogoutTargetsError(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     tcpTestPortal,
		TargetNqn:  testTarget,
		TrType:     "tcp",
		AdrFam:     "fibre-channel",
		SubType:    "ipv4",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "tcp",
	}
	GONVMEMock.InduceLogoutError = true
	err := c.NVMeTCPConnect(tgt, false)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeDisconnect(tgt)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockNVMeTCPLoginTargetsError(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     tcpTestPortal,
		TargetNqn:  testTarget,
		TrType:     "tcp",
		AdrFam:     "fibre-channel",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "tcp",
	}
	GONVMEMock.InduceTCPLoginError = true
	err := c.NVMeTCPConnect(tgt, false)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockNVMeFCLoginTargetsError(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     fcTestPortal,
		TargetNqn:  testTarget,
		TrType:     "fc",
		AdrFam:     "fibre-channel",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "fc",
		HostAdr:    hostAddress,
	}
	GONVMEMock.InduceFCLoginError = true
	err := c.NVMeFCConnect(tgt, false)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockGetSessions(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
	// check without induced error
	data, err := c.GetSessions()
	if len(data) == 0 || len(data[0].Target) == 0 {
		t.Error("invalid response from mock")
	}
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMockListNVMeDeviceAndNamespace(t *testing.T) {
	reset()
	var c NVMEinterface
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfNamespaceDevices] = fmt.Sprintf("%d", expected)
	c = NewMockNVMe(opts)
	// c = mock
	targets, err := c.ListNVMeDeviceAndNamespace()
	if err != nil {
		t.Error(err.Error())
	}
	if len(targets) != expected {
		t.Errorf("Expected to find %d targets, but got back %v", expected, targets)
	}
}

func TestMockListNVMeDeviceAndNamespaceError(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfNamespaceDevices] = fmt.Sprintf("%d", expected)
	c := NewMockNVMe(opts)
	GONVMEMock.InducedNVMeDeviceAndNamespaceError = true
	targets, err := c.ListNVMeDeviceAndNamespace()
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
	if len(targets) != 0 {
		t.Errorf("Expected to receive 0 targets when inducing an error. Received %v", targets)
		return
	}
}

func TestMockListNVMeNamespaceID(t *testing.T) {
	reset()
	var c NVMEinterface
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfNamespaceDevices] = fmt.Sprintf("%d", expected)
	c = NewMockNVMe(opts)
	// c = mock
	devices, _ := c.ListNVMeDeviceAndNamespace()
	targets, err := c.ListNVMeNamespaceID(devices)
	if err != nil {
		t.Error(err.Error())
	}
	if len(targets) != expected {
		t.Errorf("Expected to find %d targets, but got back %v", expected, targets)
	}
}

func TestMockListNVMeNamespaceIDError(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfNamespaceDevices] = fmt.Sprintf("%d", expected)
	c := NewMockNVMe(opts)
	devices, _ := c.ListNVMeDeviceAndNamespace()

	GONVMEMock.InducedNVMeNamespaceIDError = true
	targets, err := c.ListNVMeNamespaceID(devices)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
	if len(targets) != 0 {
		t.Errorf("Expected to receive 0 targets when inducing an error. Received %v", targets)
		return
	}
}

func TestMockGetNVMeDeviceData(t *testing.T) {
	reset()
	var c NVMEinterface
	opts := map[string]string{}
	c = NewMockNVMe(opts)
	_, _, err := c.GetNVMeDeviceData("/nvmeMock/0n1")
	if err != nil {
		t.Error(err.Error())
	}
}

func TestMockGetNVMeDeviceDataError(t *testing.T) {
	reset()
	var c NVMEinterface
	opts := map[string]string{}
	c = NewMockNVMe(opts)
	GONVMEMock.InducedNVMeDeviceDataError = true
	_, _, err := c.GetNVMeDeviceData("/nvmeMock/0n1")
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockGetSessionsError(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
	// check with induced error
	GONVMEMock.InduceGetSessionsError = true
	_, err := c.GetSessions()
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestSessionParserParse(t *testing.T) {
	sp := &sessionParser{}
	fileErrMsg := "can't read file with test data"

	// test valid data
	data, err := os.ReadFile("testdata/session_info_valid")
	if err != nil {
		t.Error(fileErrMsg)
	}
	sessions := sp.Parse(data)
	if len(sessions) != 2 {
		t.Error("unexpected results count")
	}
	for i, session := range sessions {
		if i == 0 {
			compareStr(t, session.Target, "nqn.1988-11.com.dell.mock:00:e6e2d5b871f1403E169D")
			compareStr(t, session.Portal, "10.230.1.1:4420")
			compareStr(t, string(session.NVMESessionState), string(NVMESessionStateLive))
			compareStr(t, string(session.NVMETransportName), string(NVMETransportNameTCP))
		} else {
			compareStr(t, session.Target, "nqn.1988-11.com.dell.mock:00:e6e2d5b871f1403E169D")
			compareStr(t, session.Portal, "10.230.1.2:4420")
			compareStr(t, string(session.NVMESessionState), string(NVMESessionStateDeleting))
			compareStr(t, string(session.NVMETransportName), string(NVMETransportNameTCP))
		}
	}

	// test invalid data parsing
	data, err = os.ReadFile("testdata/session_info_invalid")
	if err != nil {
		t.Error(fileErrMsg)
	}
	r := sp.Parse(data)
	if len(r) != 0 {
		t.Error("non empty result while parsing invalid data")
	}
}

func compareStr(t *testing.T, str1 string, str2 string) {
	if str1 != str2 {
		t.Errorf("strings are not equal: %s != %s", str1, str2)
	}
}

func TestMockDeviceRescan(t *testing.T) {
	reset()

	// Create a mock NVMe interface
	c := NewMockNVMe(map[string]string{})

	// Test successful rescan (no induced error)
	err := c.DeviceRescan("testDevice")
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Induce an error and test failure case
	GONVMEMock.InduceGetSessionsError = true
	err = c.DeviceRescan("testDevice")
	if err == nil {
		t.Error("Expected an induced error but got nil")
		return
	}
}
