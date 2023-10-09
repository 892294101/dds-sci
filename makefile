GOCMD=go
export TAG=7.1.0
export BD=$(shell date '+%b %d %Y %T')
GOBUILD=${GOCMD} build -gcflags=all="-l -N" -ldflags "-s -w -X 'github.com/892294101/dds/sci/terminal.Version=$(TAG)' -X 'github.com/892294101/dds/sci/terminal.BDate=$(BD)'"

BUILD_DIR=../build
BINARY_DIR=$(BUILD_DIR)/bin
PLUGIN_DIR=$(BUILD_DIR)/lib
LOGS_DIR=$(BUILD_DIR)/logs

BINARY_FILE=$(BINARY_DIR)/cli
BINARY_SRC=./cli.go

PLUGINS_SYSCMD_FILE=$(PLUGIN_DIR)/sys_command.so
PLUGINS_SYSCMD_SRC=./terminal/plugins/syscmd.go

PLUGINS_ADDEXTRACTCMD_FILE=$(PLUGIN_DIR)/add_command.so
PLUGINS_ADDEXTRACTCMD_SRC=./terminal/plugins/addcmd.go

PLUGINS_INFOCMD_FILE=$(PLUGIN_DIR)/info_command.so
PLUGINS_INFOCMD_SRC=./terminal/plugins/infocmd.go

PLUGINS_DELETECMD_FILE=$(PLUGIN_DIR)/delete_command.so
PLUGINS_DELETECMD_SRC=./terminal/plugins/deletecmd.go

PLUGINS_ALTERCMD_FILE=$(PLUGIN_DIR)/alter_command.so
PLUGINS_ALTERCMD_SRC=./terminal/plugins/altercmd.go

PLUGINS_STARTCMD_FILE=$(PLUGIN_DIR)/start_command.so
PLUGINS_STARTCMD_SRC=./terminal/plugins/startcmd.go

PLUGINS_STOPCMD_FILE=$(PLUGIN_DIR)/stop_command.so
PLUGINS_STOPCMD_SRC=./terminal/plugins/stopcmd.go

PLUGINS_EDITCMD_FILE=$(PLUGIN_DIR)/edit_command.so
PLUGINS_EDITCMD_SRC=./terminal/plugins/editcmd.go

PLUGINS_KILLCMD_FILE=$(PLUGIN_DIR)/kill_command.so
PLUGINS_KILLCMD_SRC=./terminal/plugins/killcmd.go

PLUGINS_VIEWCMD_FILE=$(PLUGIN_DIR)/view_command.so
PLUGINS_VIEWCMD_SRC=./terminal/plugins/viewcmd.go

.PHONY: all clean build

all: clean build

clean:
	@if [ -f  ${BUILD_DIR} ]; then rm -rf ${BUILD_DIR}/*; fi
	mkdir -p ${BINARY_DIR} ${PLUGIN_DIR} ${LOGS_DIR}

build:
	${GOBUILD} -o ${BINARY_FILE} ${BINARY_SRC}
	${GOBUILD} -buildmode=plugin -o ${PLUGINS_SYSCMD_FILE} ${PLUGINS_SYSCMD_SRC}
	${GOBUILD} -buildmode=plugin -o ${PLUGINS_ADDEXTRACTCMD_FILE} ${PLUGINS_ADDEXTRACTCMD_SRC}
	${GOBUILD} -buildmode=plugin -o ${PLUGINS_INFOCMD_FILE} ${PLUGINS_INFOCMD_SRC}
	${GOBUILD} -buildmode=plugin -o ${PLUGINS_DELETECMD_FILE} ${PLUGINS_DELETECMD_SRC}
	${GOBUILD} -buildmode=plugin -o ${PLUGINS_ALTERCMD_FILE} ${PLUGINS_ALTERCMD_SRC}
	${GOBUILD} -buildmode=plugin -o ${PLUGINS_ALTERCMD_FILE} ${PLUGINS_ALTERCMD_SRC}
	${GOBUILD} -buildmode=plugin -o ${PLUGINS_STARTCMD_FILE} ${PLUGINS_STARTCMD_SRC}
	${GOBUILD} -buildmode=plugin -o ${PLUGINS_STOPCMD_FILE} ${PLUGINS_STOPCMD_SRC}
	${GOBUILD} -buildmode=plugin -o ${PLUGINS_EDITCMD_FILE} ${PLUGINS_EDITCMD_SRC}
	${GOBUILD} -buildmode=plugin -o ${PLUGINS_KILLCMD_FILE} ${PLUGINS_KILLCMD_SRC}
	${GOBUILD} -buildmode=plugin -o ${PLUGINS_VIEWCMD_FILE} ${PLUGINS_VIEWCMD_SRC}
