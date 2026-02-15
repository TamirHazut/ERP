// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var auth_v1_user_pb = require('../../auth/v1/user_pb.js');
var infra_v1_infra_pb = require('../../infra/v1/infra_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var google_protobuf_struct_pb = require('google-protobuf/google/protobuf/struct_pb.js');
var tagger_tagger_pb = require('../../tagger/tagger_pb.js');

function serialize_auth_v1_CreateUserRequest(arg) {
  if (!(arg instanceof auth_v1_user_pb.CreateUserRequest)) {
    throw new Error('Expected argument of type auth.v1.CreateUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_CreateUserRequest(buffer_arg) {
  return auth_v1_user_pb.CreateUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_CreateUserResponse(arg) {
  if (!(arg instanceof auth_v1_user_pb.CreateUserResponse)) {
    throw new Error('Expected argument of type auth.v1.CreateUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_CreateUserResponse(buffer_arg) {
  return auth_v1_user_pb.CreateUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_DeleteUserRequest(arg) {
  if (!(arg instanceof auth_v1_user_pb.DeleteUserRequest)) {
    throw new Error('Expected argument of type auth.v1.DeleteUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_DeleteUserRequest(buffer_arg) {
  return auth_v1_user_pb.DeleteUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_DeleteUserResponse(arg) {
  if (!(arg instanceof auth_v1_user_pb.DeleteUserResponse)) {
    throw new Error('Expected argument of type auth.v1.DeleteUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_DeleteUserResponse(buffer_arg) {
  return auth_v1_user_pb.DeleteUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_GetUserRequest(arg) {
  if (!(arg instanceof auth_v1_user_pb.GetUserRequest)) {
    throw new Error('Expected argument of type auth.v1.GetUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_GetUserRequest(buffer_arg) {
  return auth_v1_user_pb.GetUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_ListUsersRequest(arg) {
  if (!(arg instanceof auth_v1_user_pb.ListUsersRequest)) {
    throw new Error('Expected argument of type auth.v1.ListUsersRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_ListUsersRequest(buffer_arg) {
  return auth_v1_user_pb.ListUsersRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_ListUsersResponse(arg) {
  if (!(arg instanceof auth_v1_user_pb.ListUsersResponse)) {
    throw new Error('Expected argument of type auth.v1.ListUsersResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_ListUsersResponse(buffer_arg) {
  return auth_v1_user_pb.ListUsersResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_UpdateUserRequest(arg) {
  if (!(arg instanceof auth_v1_user_pb.UpdateUserRequest)) {
    throw new Error('Expected argument of type auth.v1.UpdateUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_UpdateUserRequest(buffer_arg) {
  return auth_v1_user_pb.UpdateUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_UpdateUserResponse(arg) {
  if (!(arg instanceof auth_v1_user_pb.UpdateUserResponse)) {
    throw new Error('Expected argument of type auth.v1.UpdateUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_UpdateUserResponse(buffer_arg) {
  return auth_v1_user_pb.UpdateUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_User(arg) {
  if (!(arg instanceof auth_v1_user_pb.User)) {
    throw new Error('Expected argument of type auth.v1.User');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_User(buffer_arg) {
  return auth_v1_user_pb.User.deserializeBinary(new Uint8Array(buffer_arg));
}


var UserServiceService = exports.UserServiceService = {
  // CRUD
createUser: {
    path: '/auth.v1.UserService/CreateUser',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_user_pb.CreateUserRequest,
    responseType: auth_v1_user_pb.CreateUserResponse,
    requestSerialize: serialize_auth_v1_CreateUserRequest,
    requestDeserialize: deserialize_auth_v1_CreateUserRequest,
    responseSerialize: serialize_auth_v1_CreateUserResponse,
    responseDeserialize: deserialize_auth_v1_CreateUserResponse,
  },
  getUser: {
    path: '/auth.v1.UserService/GetUser',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_user_pb.GetUserRequest,
    responseType: auth_v1_user_pb.User,
    requestSerialize: serialize_auth_v1_GetUserRequest,
    requestDeserialize: deserialize_auth_v1_GetUserRequest,
    responseSerialize: serialize_auth_v1_User,
    responseDeserialize: deserialize_auth_v1_User,
  },
  listUsers: {
    path: '/auth.v1.UserService/ListUsers',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_user_pb.ListUsersRequest,
    responseType: auth_v1_user_pb.ListUsersResponse,
    requestSerialize: serialize_auth_v1_ListUsersRequest,
    requestDeserialize: deserialize_auth_v1_ListUsersRequest,
    responseSerialize: serialize_auth_v1_ListUsersResponse,
    responseDeserialize: deserialize_auth_v1_ListUsersResponse,
  },
  updateUser: {
    path: '/auth.v1.UserService/UpdateUser',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_user_pb.UpdateUserRequest,
    responseType: auth_v1_user_pb.UpdateUserResponse,
    requestSerialize: serialize_auth_v1_UpdateUserRequest,
    requestDeserialize: deserialize_auth_v1_UpdateUserRequest,
    responseSerialize: serialize_auth_v1_UpdateUserResponse,
    responseDeserialize: deserialize_auth_v1_UpdateUserResponse,
  },
  deleteUser: {
    path: '/auth.v1.UserService/DeleteUser',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_user_pb.DeleteUserRequest,
    responseType: auth_v1_user_pb.DeleteUserResponse,
    requestSerialize: serialize_auth_v1_DeleteUserRequest,
    requestDeserialize: deserialize_auth_v1_DeleteUserRequest,
    responseSerialize: serialize_auth_v1_DeleteUserResponse,
    responseDeserialize: deserialize_auth_v1_DeleteUserResponse,
  },
};

exports.UserServiceClient = grpc.makeGenericClientConstructor(UserServiceService, 'UserService');
