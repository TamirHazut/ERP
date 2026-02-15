// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var auth_v1_auth_pb = require('../../auth/v1/auth_pb.js');
var infra_v1_infra_pb = require('../../infra/v1/infra_pb.js');

function serialize_auth_v1_LoginRequest(arg) {
  if (!(arg instanceof auth_v1_auth_pb.LoginRequest)) {
    throw new Error('Expected argument of type auth.v1.LoginRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_LoginRequest(buffer_arg) {
  return auth_v1_auth_pb.LoginRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_LogoutRequest(arg) {
  if (!(arg instanceof auth_v1_auth_pb.LogoutRequest)) {
    throw new Error('Expected argument of type auth.v1.LogoutRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_LogoutRequest(buffer_arg) {
  return auth_v1_auth_pb.LogoutRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_LogoutResponse(arg) {
  if (!(arg instanceof auth_v1_auth_pb.LogoutResponse)) {
    throw new Error('Expected argument of type auth.v1.LogoutResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_LogoutResponse(buffer_arg) {
  return auth_v1_auth_pb.LogoutResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_RefreshTokenRequest(arg) {
  if (!(arg instanceof auth_v1_auth_pb.RefreshTokenRequest)) {
    throw new Error('Expected argument of type auth.v1.RefreshTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_RefreshTokenRequest(buffer_arg) {
  return auth_v1_auth_pb.RefreshTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_RevokeAllTenantTokensRequest(arg) {
  if (!(arg instanceof auth_v1_auth_pb.RevokeAllTenantTokensRequest)) {
    throw new Error('Expected argument of type auth.v1.RevokeAllTenantTokensRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_RevokeAllTenantTokensRequest(buffer_arg) {
  return auth_v1_auth_pb.RevokeAllTenantTokensRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_RevokeAllTenantTokensResponse(arg) {
  if (!(arg instanceof auth_v1_auth_pb.RevokeAllTenantTokensResponse)) {
    throw new Error('Expected argument of type auth.v1.RevokeAllTenantTokensResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_RevokeAllTenantTokensResponse(buffer_arg) {
  return auth_v1_auth_pb.RevokeAllTenantTokensResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_RevokeTokenRequest(arg) {
  if (!(arg instanceof auth_v1_auth_pb.RevokeTokenRequest)) {
    throw new Error('Expected argument of type auth.v1.RevokeTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_RevokeTokenRequest(buffer_arg) {
  return auth_v1_auth_pb.RevokeTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_RevokeTokenResponse(arg) {
  if (!(arg instanceof auth_v1_auth_pb.RevokeTokenResponse)) {
    throw new Error('Expected argument of type auth.v1.RevokeTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_RevokeTokenResponse(buffer_arg) {
  return auth_v1_auth_pb.RevokeTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_Tokens(arg) {
  if (!(arg instanceof auth_v1_auth_pb.Tokens)) {
    throw new Error('Expected argument of type auth.v1.Tokens');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_Tokens(buffer_arg) {
  return auth_v1_auth_pb.Tokens.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_VerifyTokenRequest(arg) {
  if (!(arg instanceof auth_v1_auth_pb.VerifyTokenRequest)) {
    throw new Error('Expected argument of type auth.v1.VerifyTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_VerifyTokenRequest(buffer_arg) {
  return auth_v1_auth_pb.VerifyTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_VerifyTokenResponse(arg) {
  if (!(arg instanceof auth_v1_auth_pb.VerifyTokenResponse)) {
    throw new Error('Expected argument of type auth.v1.VerifyTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_VerifyTokenResponse(buffer_arg) {
  return auth_v1_auth_pb.VerifyTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var AuthServiceService = exports.AuthServiceService = {
  // Authentication - Login + Logout
login: {
    path: '/auth.v1.AuthService/Login',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_auth_pb.LoginRequest,
    responseType: auth_v1_auth_pb.Tokens,
    requestSerialize: serialize_auth_v1_LoginRequest,
    requestDeserialize: deserialize_auth_v1_LoginRequest,
    responseSerialize: serialize_auth_v1_Tokens,
    responseDeserialize: deserialize_auth_v1_Tokens,
  },
  logout: {
    path: '/auth.v1.AuthService/Logout',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_auth_pb.LogoutRequest,
    responseType: auth_v1_auth_pb.LogoutResponse,
    requestSerialize: serialize_auth_v1_LogoutRequest,
    requestDeserialize: deserialize_auth_v1_LogoutRequest,
    responseSerialize: serialize_auth_v1_LogoutResponse,
    responseDeserialize: deserialize_auth_v1_LogoutResponse,
  },
  // Access + Refresh Tokens
verifyToken: {
    path: '/auth.v1.AuthService/VerifyToken',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_auth_pb.VerifyTokenRequest,
    responseType: auth_v1_auth_pb.VerifyTokenResponse,
    requestSerialize: serialize_auth_v1_VerifyTokenRequest,
    requestDeserialize: deserialize_auth_v1_VerifyTokenRequest,
    responseSerialize: serialize_auth_v1_VerifyTokenResponse,
    responseDeserialize: deserialize_auth_v1_VerifyTokenResponse,
  },
  refreshToken: {
    path: '/auth.v1.AuthService/RefreshToken',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_auth_pb.RefreshTokenRequest,
    responseType: auth_v1_auth_pb.Tokens,
    requestSerialize: serialize_auth_v1_RefreshTokenRequest,
    requestDeserialize: deserialize_auth_v1_RefreshTokenRequest,
    responseSerialize: serialize_auth_v1_Tokens,
    responseDeserialize: deserialize_auth_v1_Tokens,
  },
  revokeToken: {
    path: '/auth.v1.AuthService/RevokeToken',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_auth_pb.RevokeTokenRequest,
    responseType: auth_v1_auth_pb.RevokeTokenResponse,
    requestSerialize: serialize_auth_v1_RevokeTokenRequest,
    requestDeserialize: deserialize_auth_v1_RevokeTokenRequest,
    responseSerialize: serialize_auth_v1_RevokeTokenResponse,
    responseDeserialize: deserialize_auth_v1_RevokeTokenResponse,
  },
  // Tenant-level token management
revokeAllTenantTokens: {
    path: '/auth.v1.AuthService/RevokeAllTenantTokens',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_auth_pb.RevokeAllTenantTokensRequest,
    responseType: auth_v1_auth_pb.RevokeAllTenantTokensResponse,
    requestSerialize: serialize_auth_v1_RevokeAllTenantTokensRequest,
    requestDeserialize: deserialize_auth_v1_RevokeAllTenantTokensRequest,
    responseSerialize: serialize_auth_v1_RevokeAllTenantTokensResponse,
    responseDeserialize: deserialize_auth_v1_RevokeAllTenantTokensResponse,
  },
};

exports.AuthServiceClient = grpc.makeGenericClientConstructor(AuthServiceService, 'AuthService');
