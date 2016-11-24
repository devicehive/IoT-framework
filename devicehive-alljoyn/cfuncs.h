/*
  DeviceHive IoT-Framework business logic

  Copyright (C) 2016 DataArt

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at
 
      http://www.apache.org/licenses/LICENSE-2.0
 
  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/ 

#include <stdio.h>
#include <aj_debug.h>
#include <aj_guid.h>
#include <aj_creds.h>
#include <aj_peer.h>
#include <aj_link_timeout.h>
#include "alljoyn.h"

uint32_t Get_AJ_Message_msgId();
uint32_t Get_AJ_Message_bodyLen();
const char * Get_AJ_Message_signature();
const char * Get_AJ_Message_objPath();
const char * Get_AJ_Message_iface();
const char * Get_AJ_Message_member();
const char * Get_AJ_Message_destination();
const char* AJ_StatusText(AJ_Status status);

const char* MyTranslator(uint32_t descId, const char* lang) ;

AJ_Message * Get_AJ_ReplyMessage();

AJ_Message * Get_AJ_Message();
//void InitNotificationContent();
AJ_Status AJNS_Producer_Start();
void SendNotification(uint16_t messageType, char * lang, char * msg);
AJ_BusAttachment * Get_AJ_BusAttachment();
AJ_Object * Allocate_AJ_Object_Array(uint32_t array_size);
void * Create_AJ_Object(uint32_t index, AJ_Object * array, char* path, AJ_InterfaceDescription* interfaces, uint8_t flags, void* context);

void * Get_Session_Opts();
void * Get_Arg();
AJ_Status AJ_MarshalArgs_cgo(AJ_Message* msg, char * a, char * b, char * c, char * d);

int UnmarshalPort();
typedef void * (*AboutPropGetter)(const char* name, const char* language);

void free (void *__ptr);
AJ_Status MarshalArg(AJ_Message* msg, char * sig, void * value);
AJ_Status AJ_DeliverMsg(AJ_Message* msg);
AJ_Status AJ_MarshalSignal_cgo(AJ_Message* msg, uint32_t msgId, uint32_t sessionId, uint8_t flags, uint32_t ttl);
AJ_Status UnmarshalJoinSessionArgs(AJ_Message* msg, uint16_t * port, uint32_t * sessionId);
AJ_Status UnmarshalLostSessionArgs(AJ_Message* msg, uint32_t * sessionId, uint32_t * reason);

void SetProperty(char* key, void * value);
void * GetProperty(char* key);

AJ_Status MyAboutPropGetter(AJ_Message* reply, const char* language);
void AJ_RegisterDescriptionLanguages(const char* const* languages);

char ** getLanguages();
//void AJ_PrintXMLWithDescriptions(const AJ_Object* objs, const char* languageTag);
