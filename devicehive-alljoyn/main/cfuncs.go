package main

/*
#include <stdio.h>
#include <aj_debug.h>
#include <aj_guid.h>
#include <aj_creds.h>
#include <aj_nvram.h>
#include "alljoyn.h"

AJ_BusAttachment c_bus;
AJ_Message c_message;
AJ_Message c_reply;
void * c_propGetter;
AJ_Arg c_arg;
AJ_SessionOpts session_opts = { AJ_SESSION_TRAFFIC_MESSAGES, AJ_SESSION_PROXIMITY_ANY, AJ_TRANSPORT_ANY, TRUE };

void * Get_Session_Opts() {
	return &session_opts;
}


void * Get_Arg() {
	return &c_arg;
}

AJ_Status AJ_MarshalArgs_cgo(AJ_Message* msg, char * a, char * b, char * c, char * d) {
	return AJ_MarshalArgs(msg, a, b, c, d);
}

uint32_t Get_AJ_Message_msgId() {
	return c_message.msgId;
}

uint32_t Get_AJ_Message_bodyLen() {
	return c_message.hdr->bodyLen;
}

const char * Get_AJ_Message_signature() {
	return c_message.signature;
}

const char * Get_AJ_Message_objPath() {
   return c_message.objPath;
}

const char * Get_AJ_Message_iface() {
   return c_message.iface;
}

const char * Get_AJ_Message_member() {
   return c_message.member;
}

const char * Get_AJ_Message_destination() {
   return c_message.destination;
}

void * Get_AJ_ReplyMessage() {
	return &c_reply;
}

void * Get_AJ_Message() {
	return &c_message;
}
void * Get_AJ_BusAttachment() {
	return &c_bus;
}

void * Allocate_AJ_Object_Array(uint32_t array_size) {
	return AJ_Malloc(sizeof(AJ_Object)*array_size);
}

void * Create_AJ_Object(uint32_t index, void * array, char* path, AJ_InterfaceDescription* interfaces, uint8_t flags, void* context) {
	AJ_Object * obj = array + index * sizeof(AJ_Object);
	obj->path = path;
	obj->interfaces = interfaces;
	obj->flags = flags;
	obj->context = context;
	return obj;
}

AJ_Status MyAboutPropGetter_cgo(AJ_Message* reply, const char* language) {
	printf("C.MyAboutPropGetter_cgo() called\n");
	return MyAboutPropGetter(reply, language);
}

AJ_Status MyRegisterConfigObject_cgo() {
	// TODO just a test
	AJ_GUID localGuid = {0xbc, 0xd3, 0x1d, 0xf4, 0xfd, 0x21, 0x59, 0xe3, 0xc8, 0x45, 0x31, 0x23, 0xb4, 0xa0, 0x01, 0x8e};
	AJ_Status status = AJ_ERR_FAILURE;
	AJ_NV_DATASET* handle = AJ_NVRAM_Open(AJ_LOCAL_GUID_NV_ID, "w", sizeof(AJ_GUID));
    if (handle) {
        if (sizeof(AJ_GUID) == AJ_NVRAM_Write(&localGuid, sizeof(AJ_GUID), handle)) {
            status = AJ_OK;
			printf("****AJ_GUID has written****\n");
        }
        status = AJ_NVRAM_Close(handle);
    }


	static const char* const AJSVC_ConfigInterface[] = {
    "$org.alljoyn.Config",
    "@Version>q",
    "?GetConfigurations <s >a{sv}",
    NULL
	};
	static const AJ_InterfaceDescription AJSVC_ConfigInterfaces[] = {
	    AJ_PropertiesIface,
	    AJSVC_ConfigInterface,
	    NULL
	};

	static AJ_Object AJCFG_ObjectList[] = {
    { "/Config", AJSVC_ConfigInterfaces, AJ_OBJ_FLAG_ANNOUNCED },
    { NULL }
	};
    return AJ_RegisterObjectList(AJCFG_ObjectList, 3);
}

int UnmarshalPort() {
	uint16_t port;
	char* joiner;
	uint32_t sessionId;

	AJ_UnmarshalArgs(&c_message, "qus", &port, &sessionId, &joiner);
	return port;
}
*/
import "C"
