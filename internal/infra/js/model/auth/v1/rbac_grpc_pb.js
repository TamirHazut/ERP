// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var auth_v1_rbac_pb = require('../../auth/v1/rbac_pb.js');
var infra_v1_infra_pb = require('../../infra/v1/infra_pb.js');
var auth_v1_role_pb = require('../../auth/v1/role_pb.js');
var auth_v1_permission_pb = require('../../auth/v1/permission_pb.js');

function serialize_auth_v1_CheckPermissionsRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.CheckPermissionsRequest)) {
    throw new Error('Expected argument of type auth.v1.CheckPermissionsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_CheckPermissionsRequest(buffer_arg) {
  return auth_v1_rbac_pb.CheckPermissionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_CheckPermissionsResponse(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.CheckPermissionsResponse)) {
    throw new Error('Expected argument of type auth.v1.CheckPermissionsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_CheckPermissionsResponse(buffer_arg) {
  return auth_v1_rbac_pb.CheckPermissionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_CreateRoleRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.CreateRoleRequest)) {
    throw new Error('Expected argument of type auth.v1.CreateRoleRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_CreateRoleRequest(buffer_arg) {
  return auth_v1_rbac_pb.CreateRoleRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_CreateRoleResponse(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.CreateRoleResponse)) {
    throw new Error('Expected argument of type auth.v1.CreateRoleResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_CreateRoleResponse(buffer_arg) {
  return auth_v1_rbac_pb.CreateRoleResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_DeleteRoleRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.DeleteRoleRequest)) {
    throw new Error('Expected argument of type auth.v1.DeleteRoleRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_DeleteRoleRequest(buffer_arg) {
  return auth_v1_rbac_pb.DeleteRoleRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_GetPermissionRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.GetPermissionRequest)) {
    throw new Error('Expected argument of type auth.v1.GetPermissionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_GetPermissionRequest(buffer_arg) {
  return auth_v1_rbac_pb.GetPermissionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_GetRoleRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.GetRoleRequest)) {
    throw new Error('Expected argument of type auth.v1.GetRoleRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_GetRoleRequest(buffer_arg) {
  return auth_v1_rbac_pb.GetRoleRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_GetUserPermissionsRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.GetUserPermissionsRequest)) {
    throw new Error('Expected argument of type auth.v1.GetUserPermissionsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_GetUserPermissionsRequest(buffer_arg) {
  return auth_v1_rbac_pb.GetUserPermissionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_GetUserPermissionsResponse(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.GetUserPermissionsResponse)) {
    throw new Error('Expected argument of type auth.v1.GetUserPermissionsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_GetUserPermissionsResponse(buffer_arg) {
  return auth_v1_rbac_pb.GetUserPermissionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_GetUserRolesRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.GetUserRolesRequest)) {
    throw new Error('Expected argument of type auth.v1.GetUserRolesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_GetUserRolesRequest(buffer_arg) {
  return auth_v1_rbac_pb.GetUserRolesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_GetUserRolesResponse(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.GetUserRolesResponse)) {
    throw new Error('Expected argument of type auth.v1.GetUserRolesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_GetUserRolesResponse(buffer_arg) {
  return auth_v1_rbac_pb.GetUserRolesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_HasPermissionRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.HasPermissionRequest)) {
    throw new Error('Expected argument of type auth.v1.HasPermissionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_HasPermissionRequest(buffer_arg) {
  return auth_v1_rbac_pb.HasPermissionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_HasPermissionResponse(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.HasPermissionResponse)) {
    throw new Error('Expected argument of type auth.v1.HasPermissionResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_HasPermissionResponse(buffer_arg) {
  return auth_v1_rbac_pb.HasPermissionResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_IsSystemTenantUserRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.IsSystemTenantUserRequest)) {
    throw new Error('Expected argument of type auth.v1.IsSystemTenantUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_IsSystemTenantUserRequest(buffer_arg) {
  return auth_v1_rbac_pb.IsSystemTenantUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_IsSystemTenantUserResponse(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.IsSystemTenantUserResponse)) {
    throw new Error('Expected argument of type auth.v1.IsSystemTenantUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_IsSystemTenantUserResponse(buffer_arg) {
  return auth_v1_rbac_pb.IsSystemTenantUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_ListPermissionsRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.ListPermissionsRequest)) {
    throw new Error('Expected argument of type auth.v1.ListPermissionsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_ListPermissionsRequest(buffer_arg) {
  return auth_v1_rbac_pb.ListPermissionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_ListPermissionsResponse(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.ListPermissionsResponse)) {
    throw new Error('Expected argument of type auth.v1.ListPermissionsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_ListPermissionsResponse(buffer_arg) {
  return auth_v1_rbac_pb.ListPermissionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_ListRolesRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.ListRolesRequest)) {
    throw new Error('Expected argument of type auth.v1.ListRolesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_ListRolesRequest(buffer_arg) {
  return auth_v1_rbac_pb.ListRolesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_ListRolesResponse(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.ListRolesResponse)) {
    throw new Error('Expected argument of type auth.v1.ListRolesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_ListRolesResponse(buffer_arg) {
  return auth_v1_rbac_pb.ListRolesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_Permission(arg) {
  if (!(arg instanceof auth_v1_permission_pb.Permission)) {
    throw new Error('Expected argument of type auth.v1.Permission');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_Permission(buffer_arg) {
  return auth_v1_permission_pb.Permission.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_Role(arg) {
  if (!(arg instanceof auth_v1_role_pb.Role)) {
    throw new Error('Expected argument of type auth.v1.Role');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_Role(buffer_arg) {
  return auth_v1_role_pb.Role.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v1_UpdateRoleRequest(arg) {
  if (!(arg instanceof auth_v1_rbac_pb.UpdateRoleRequest)) {
    throw new Error('Expected argument of type auth.v1.UpdateRoleRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v1_UpdateRoleRequest(buffer_arg) {
  return auth_v1_rbac_pb.UpdateRoleRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_infra_v1_Response(arg) {
  if (!(arg instanceof infra_v1_infra_pb.Response)) {
    throw new Error('Expected argument of type infra.v1.Response');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_infra_v1_Response(buffer_arg) {
  return infra_v1_infra_pb.Response.deserializeBinary(new Uint8Array(buffer_arg));
}


// ============================================================================
// Dedicated Service Definitions
// ============================================================================
//
// RoleService provides role management operations
var RoleServiceService = exports.RoleServiceService = {
  createRole: {
    path: '/auth.v1.RoleService/CreateRole',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.CreateRoleRequest,
    responseType: auth_v1_rbac_pb.CreateRoleResponse,
    requestSerialize: serialize_auth_v1_CreateRoleRequest,
    requestDeserialize: deserialize_auth_v1_CreateRoleRequest,
    responseSerialize: serialize_auth_v1_CreateRoleResponse,
    responseDeserialize: deserialize_auth_v1_CreateRoleResponse,
  },
  updateRole: {
    path: '/auth.v1.RoleService/UpdateRole',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.UpdateRoleRequest,
    responseType: infra_v1_infra_pb.Response,
    requestSerialize: serialize_auth_v1_UpdateRoleRequest,
    requestDeserialize: deserialize_auth_v1_UpdateRoleRequest,
    responseSerialize: serialize_infra_v1_Response,
    responseDeserialize: deserialize_infra_v1_Response,
  },
  getRole: {
    path: '/auth.v1.RoleService/GetRole',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.GetRoleRequest,
    responseType: auth_v1_role_pb.Role,
    requestSerialize: serialize_auth_v1_GetRoleRequest,
    requestDeserialize: deserialize_auth_v1_GetRoleRequest,
    responseSerialize: serialize_auth_v1_Role,
    responseDeserialize: deserialize_auth_v1_Role,
  },
  listRoles: {
    path: '/auth.v1.RoleService/ListRoles',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.ListRolesRequest,
    responseType: auth_v1_rbac_pb.ListRolesResponse,
    requestSerialize: serialize_auth_v1_ListRolesRequest,
    requestDeserialize: deserialize_auth_v1_ListRolesRequest,
    responseSerialize: serialize_auth_v1_ListRolesResponse,
    responseDeserialize: deserialize_auth_v1_ListRolesResponse,
  },
  deleteRole: {
    path: '/auth.v1.RoleService/DeleteRole',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.DeleteRoleRequest,
    responseType: infra_v1_infra_pb.Response,
    requestSerialize: serialize_auth_v1_DeleteRoleRequest,
    requestDeserialize: deserialize_auth_v1_DeleteRoleRequest,
    responseSerialize: serialize_infra_v1_Response,
    responseDeserialize: deserialize_infra_v1_Response,
  },
};

exports.RoleServiceClient = grpc.makeGenericClientConstructor(RoleServiceService, 'RoleService');
// PermissionService provides permission read operations backed by the code-defined registry
var PermissionServiceService = exports.PermissionServiceService = {
  getPermission: {
    path: '/auth.v1.PermissionService/GetPermission',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.GetPermissionRequest,
    responseType: auth_v1_permission_pb.Permission,
    requestSerialize: serialize_auth_v1_GetPermissionRequest,
    requestDeserialize: deserialize_auth_v1_GetPermissionRequest,
    responseSerialize: serialize_auth_v1_Permission,
    responseDeserialize: deserialize_auth_v1_Permission,
  },
  listPermissions: {
    path: '/auth.v1.PermissionService/ListPermissions',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.ListPermissionsRequest,
    responseType: auth_v1_rbac_pb.ListPermissionsResponse,
    requestSerialize: serialize_auth_v1_ListPermissionsRequest,
    requestDeserialize: deserialize_auth_v1_ListPermissionsRequest,
    responseSerialize: serialize_auth_v1_ListPermissionsResponse,
    responseDeserialize: deserialize_auth_v1_ListPermissionsResponse,
  },
};

exports.PermissionServiceClient = grpc.makeGenericClientConstructor(PermissionServiceService, 'PermissionService');
// VerificationService provides permission and role verification operations
var VerificationServiceService = exports.VerificationServiceService = {
  checkPermissions: {
    path: '/auth.v1.VerificationService/CheckPermissions',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.CheckPermissionsRequest,
    responseType: auth_v1_rbac_pb.CheckPermissionsResponse,
    requestSerialize: serialize_auth_v1_CheckPermissionsRequest,
    requestDeserialize: deserialize_auth_v1_CheckPermissionsRequest,
    responseSerialize: serialize_auth_v1_CheckPermissionsResponse,
    responseDeserialize: deserialize_auth_v1_CheckPermissionsResponse,
  },
  hasPermission: {
    path: '/auth.v1.VerificationService/HasPermission',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.HasPermissionRequest,
    responseType: auth_v1_rbac_pb.HasPermissionResponse,
    requestSerialize: serialize_auth_v1_HasPermissionRequest,
    requestDeserialize: deserialize_auth_v1_HasPermissionRequest,
    responseSerialize: serialize_auth_v1_HasPermissionResponse,
    responseDeserialize: deserialize_auth_v1_HasPermissionResponse,
  },
  getUserPermissions: {
    path: '/auth.v1.VerificationService/GetUserPermissions',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.GetUserPermissionsRequest,
    responseType: auth_v1_rbac_pb.GetUserPermissionsResponse,
    requestSerialize: serialize_auth_v1_GetUserPermissionsRequest,
    requestDeserialize: deserialize_auth_v1_GetUserPermissionsRequest,
    responseSerialize: serialize_auth_v1_GetUserPermissionsResponse,
    responseDeserialize: deserialize_auth_v1_GetUserPermissionsResponse,
  },
  getUserRoles: {
    path: '/auth.v1.VerificationService/GetUserRoles',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.GetUserRolesRequest,
    responseType: auth_v1_rbac_pb.GetUserRolesResponse,
    requestSerialize: serialize_auth_v1_GetUserRolesRequest,
    requestDeserialize: deserialize_auth_v1_GetUserRolesRequest,
    responseSerialize: serialize_auth_v1_GetUserRolesResponse,
    responseDeserialize: deserialize_auth_v1_GetUserRolesResponse,
  },
  isSystemTenantUser: {
    path: '/auth.v1.VerificationService/IsSystemTenantUser',
    requestStream: false,
    responseStream: false,
    requestType: auth_v1_rbac_pb.IsSystemTenantUserRequest,
    responseType: auth_v1_rbac_pb.IsSystemTenantUserResponse,
    requestSerialize: serialize_auth_v1_IsSystemTenantUserRequest,
    requestDeserialize: deserialize_auth_v1_IsSystemTenantUserRequest,
    responseSerialize: serialize_auth_v1_IsSystemTenantUserResponse,
    responseDeserialize: deserialize_auth_v1_IsSystemTenantUserResponse,
  },
};

exports.VerificationServiceClient = grpc.makeGenericClientConstructor(VerificationServiceService, 'VerificationService');
