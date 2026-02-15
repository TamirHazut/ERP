// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var config_v1_config_pb = require('../../config/v1/config_pb.js');
var google_protobuf_struct_pb = require('google-protobuf/google/protobuf/struct_pb.js');

function serialize_config_v1_ConfigRequest(arg) {
  if (!(arg instanceof config_v1_config_pb.ConfigRequest)) {
    throw new Error('Expected argument of type config.v1.ConfigRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_config_v1_ConfigRequest(buffer_arg) {
  return config_v1_config_pb.ConfigRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_config_v1_ConfigResponse(arg) {
  if (!(arg instanceof config_v1_config_pb.ConfigResponse)) {
    throw new Error('Expected argument of type config.v1.ConfigResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_config_v1_ConfigResponse(buffer_arg) {
  return config_v1_config_pb.ConfigResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_config_v1_EnvRequest(arg) {
  if (!(arg instanceof config_v1_config_pb.EnvRequest)) {
    throw new Error('Expected argument of type config.v1.EnvRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_config_v1_EnvRequest(buffer_arg) {
  return config_v1_config_pb.EnvRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_config_v1_EnvResponse(arg) {
  if (!(arg instanceof config_v1_config_pb.EnvResponse)) {
    throw new Error('Expected argument of type config.v1.EnvResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_config_v1_EnvResponse(buffer_arg) {
  return config_v1_config_pb.EnvResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_config_v1_FeatureFlagRequest(arg) {
  if (!(arg instanceof config_v1_config_pb.FeatureFlagRequest)) {
    throw new Error('Expected argument of type config.v1.FeatureFlagRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_config_v1_FeatureFlagRequest(buffer_arg) {
  return config_v1_config_pb.FeatureFlagRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_config_v1_FeatureFlagResponse(arg) {
  if (!(arg instanceof config_v1_config_pb.FeatureFlagResponse)) {
    throw new Error('Expected argument of type config.v1.FeatureFlagResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_config_v1_FeatureFlagResponse(buffer_arg) {
  return config_v1_config_pb.FeatureFlagResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var ConfigServiceService = exports.ConfigServiceService = {
  getConfig: {
    path: '/config.v1.ConfigService/GetConfig',
    requestStream: false,
    responseStream: false,
    requestType: config_v1_config_pb.ConfigRequest,
    responseType: config_v1_config_pb.ConfigResponse,
    requestSerialize: serialize_config_v1_ConfigRequest,
    requestDeserialize: deserialize_config_v1_ConfigRequest,
    responseSerialize: serialize_config_v1_ConfigResponse,
    responseDeserialize: deserialize_config_v1_ConfigResponse,
  },
  getEnv: {
    path: '/config.v1.ConfigService/GetEnv',
    requestStream: false,
    responseStream: false,
    requestType: config_v1_config_pb.EnvRequest,
    responseType: config_v1_config_pb.EnvResponse,
    requestSerialize: serialize_config_v1_EnvRequest,
    requestDeserialize: deserialize_config_v1_EnvRequest,
    responseSerialize: serialize_config_v1_EnvResponse,
    responseDeserialize: deserialize_config_v1_EnvResponse,
  },
  setFeatureFlag: {
    path: '/config.v1.ConfigService/SetFeatureFlag',
    requestStream: false,
    responseStream: false,
    requestType: config_v1_config_pb.FeatureFlagRequest,
    responseType: config_v1_config_pb.FeatureFlagResponse,
    requestSerialize: serialize_config_v1_FeatureFlagRequest,
    requestDeserialize: deserialize_config_v1_FeatureFlagRequest,
    responseSerialize: serialize_config_v1_FeatureFlagResponse,
    responseDeserialize: deserialize_config_v1_FeatureFlagResponse,
  },
};

exports.ConfigServiceClient = grpc.makeGenericClientConstructor(ConfigServiceService, 'ConfigService');
