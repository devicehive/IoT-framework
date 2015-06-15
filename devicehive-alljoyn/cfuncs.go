package main

/*
#include <stdio.h>
#include <aj_debug.h>
#include <aj_guid.h>
#include <aj_creds.h>
#include "alljoyn.h"

AJ_BusAttachment c_bus;
AJ_Message c_message;
AJ_Message c_reply;
void * c_propGetter;
AJ_Arg c_arg;
AJ_MsgHeader c_msgHeader;

void * Get_MsgHeader() {
	return &c_msgHeader;
}

AJ_MsgHeader * BackupMsgHeader(AJ_MsgHeader * src) {
	return memcpy(Get_MsgHeader(), src, sizeof(AJ_MsgHeader));
}

void * Get_Arg() {
	return &c_arg;
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

AJ_Status AJ_MarshalArgs_cgo(AJ_Message* msg, char * a, char * b, char * c, char * d) {
	return AJ_MarshalArgs(msg, a, b, c, d);
}
*/
import "C"
