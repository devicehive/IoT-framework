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
	// BE CAREFULL WHEN YOU WILL IMPLEMENT OBJECT DELETION
	// YOU MUST DELETE ALL ALLOCS
	AJ_Object * obj = array + index * sizeof(AJ_Object);
	if(path) {
		char *c = AJ_Malloc(strlen(path) + 1);
		strcpy(c, path);
		obj->path = c;
	} else {
		obj->path = 0;
	}

	if(interfaces) {
		int ic = 0;
		while(interfaces[ic++]);
		AJ_InterfaceDescription *interfacescopy = AJ_Malloc(ic * sizeof(AJ_InterfaceDescription*));
		int i;
		for(i = 0; i < ic; i++) {
			if(interfaces[i]) {
				int iic = 0;
				while(interfaces[i][iic++]);
				char **newitem = AJ_Malloc(iic * sizeof(char *));
				int j;
				for(j = 0; j < iic; j++) {
					if(interfaces[i][j]) {
						char *c = AJ_Malloc(strlen(interfaces[i][j]) + 1);
						strcpy(c, interfaces[i][j]);
						newitem[j] = c;
					} else {
						newitem[j] = 0;
					}
				}
				interfacescopy[i] = (AJ_InterfaceDescription)newitem;
			} else {
				interfacescopy[i] = 0;
			}
		}
		obj->interfaces = interfacescopy;
	} else {
		obj->interfaces = 0;
	}

	obj->flags = flags;
	obj->context = context;
	return obj;
}

AJ_Status MyAboutPropGetter_cgo(AJ_Message* reply, const char* language) {
	printf("C.MyAboutPropGetter_cgo() called\n");
	return MyAboutPropGetter(reply, language);
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
