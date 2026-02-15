// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var auth_v1_tenant_pb = require('../../auth/v1/tenant_pb.js');
var infra_v1_infra_pb = require('../../infra/v1/infra_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var tagger_tagger_pb = require('../../tagger/tagger_pb.js');
var core_v1_address_pb = require('../../core/v1/address_pb.js');

function serialize_auth_v1_CreateTenantRequest(arg) {
  if (!(arg instanceof auth_v1_tenant_pb.CreateTenantRequest)) {
    throw new Error('Expected argument of type auth.v1.CreateTenantRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_CreateTenantRequest(buffer_arg) {
  return auth_v1_tenant_pb.CreateTenantRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_CreateTenantResponse(arg) {
  if (!(arg instanceof auth_v1_tenant_pb.CreateTenantResponse)) {
    throw new Error('Expected argument of type auth.v1.CreateTenantResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_CreateTenantResponse(buffer_arg) {
  return auth_v1_tenant_pb.CreateTenantResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_DeleteTenantRequest(arg) {
  if (!(arg instanceof auth_v1_tenant_pb.DeleteTenantRequest)) {
    throw new Error('Expected argument of type auth.v1.DeleteTenantRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_DeleteTenantRequest(buffer_arg) {
  return auth_v1_tenant_pb.DeleteTenantRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_DeleteTenantResponse(arg) {
  if (!(arg instanceof auth_v1_tenant_pb.DeleteTenantResponse)) {
    throw new Error('Expected argument of type auth.v1.DeleteTenantResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_DeleteTenantResponse(buffer_arg) {
  return auth_v1_tenant_pb.DeleteTenantResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_GetTenantRequest(arg) {
  if (!(arg instanceof auth_v1_tenant_pb.GetTenantRequest)) {
    throw new Error('Expected argument of type auth.v1.GetTenantRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_GetTenantRequest(buffer_arg) {
  return auth_v1_tenant_pb.GetTenantRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_ListTenantsRequest(arg) {
  if (!(arg instanceof auth_v1_tenant_pb.ListTenantsRequest)) {
    throw new Error('Expected argument of type auth.v1.ListTenantsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_ListTenantsRequest(buffer_arg) {
  return auth_v1_tenant_pb.ListTenantsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_ListTenantsResponse(arg) {
  if (!(arg instanceof auth_v1_tenant_pb.ListTenantsResponse)) {
    throw new Error('Expected argument of type auth.v1.ListTenantsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_ListTenantsResponse(buffer_arg) {
  return auth_v1_tenant_pb.ListTenantsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_Tenant(arg) {
  if (!(arg instanceof auth_v1_tenant_pb.Tenant)) {
    throw new Error('Expected argument of type auth.v1.Tenant');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_Tenant(buffer_arg) {
  return auth_v1_tenant_pb.Tenant.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_UpdateTenantRequest(arg) {
  if (!(arg instanceof auth_v1_tenant_pb.UpdateTenantRequest)) {
    throw new Error('Expected argument of type auth.v1.UpdateTenantRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_UpdateTenantRequest(buffer_arg) {
  return auth_v1_tenant_pb.UpdateTenantRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_UpdateTenantResponse(arg) {
  if (!(arg instanceof auth_v1_tenant_pb.UpdateTenantResponse)) {
    throw new Error('Expected argument of type auth.v1.UpdateTenantResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_UpdateTenantResponse(buffer_arg) {
  return auth_v1_tenant_pb.UpdateTenantResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


// =============================================================================
// Service Definition
// =============================================================================
//
var TenantServiceService = exports.TenantServiceService = {
  // CRUD
createTenant: {
    path: '/auth.v1.TenantService/CreateTenant',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_tenant_pb.CreateTenantRequest,
    responseType: auth_v1_tenant_pb.CreateTenantResponse,
    requestSerialize: serialize_auth_v1_CreateTenantRequest,
    requestDeserialize: deserialize_auth_v1_CreateTenantRequest,
    responseSerialize: serialize_auth_v1_CreateTenantResponse,
    responseDeserialize: deserialize_auth_v1_CreateTenantResponse,
  },
  getTenant: {
    path: '/auth.v1.TenantService/GetTenant',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_tenant_pb.GetTenantRequest,
    responseType: auth_v1_tenant_pb.Tenant,
    requestSerialize: serialize_auth_v1_GetTenantRequest,
    requestDeserialize: deserialize_auth_v1_GetTenantRequest,
    responseSerialize: serialize_auth_v1_Tenant,
    responseDeserialize: deserialize_auth_v1_Tenant,
  },
  listTenants: {
    path: '/auth.v1.TenantService/ListTenants',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_tenant_pb.ListTenantsRequest,
    responseType: auth_v1_tenant_pb.ListTenantsResponse,
    requestSerialize: serialize_auth_v1_ListTenantsRequest,
    requestDeserialize: deserialize_auth_v1_ListTenantsRequest,
    responseSerialize: serialize_auth_v1_ListTenantsResponse,
    responseDeserialize: deserialize_auth_v1_ListTenantsResponse,
  },
  updateTenant: {
    path: '/auth.v1.TenantService/UpdateTenant',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_tenant_pb.UpdateTenantRequest,
    responseType: auth_v1_tenant_pb.UpdateTenantResponse,
    requestSerialize: serialize_auth_v1_UpdateTenantRequest,
    requestDeserialize: deserialize_auth_v1_UpdateTenantRequest,
    responseSerialize: serialize_auth_v1_UpdateTenantResponse,
    responseDeserialize: deserialize_auth_v1_UpdateTenantResponse,
  },
  deleteTenant: {
    path: '/auth.v1.TenantService/DeleteTenant',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_tenant_pb.DeleteTenantRequest,
    responseType: auth_v1_tenant_pb.DeleteTenantResponse,
    requestSerialize: serialize_auth_v1_DeleteTenantRequest,
    requestDeserialize: deserialize_auth_v1_DeleteTenantRequest,
    responseSerialize: serialize_auth_v1_DeleteTenantResponse,
    responseDeserialize: deserialize_auth_v1_DeleteTenantResponse,
  },
};

exports.TenantServiceClient = grpc.makeGenericClientConstructor(TenantServiceService, 'TenantService');
