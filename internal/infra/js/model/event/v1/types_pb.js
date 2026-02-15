// source: event/v1/types.proto
/**
 * @fileoverview
 * @enhanceable
 * @suppress {missingRequire} reports error on implicit type usages.
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!
/* eslint-disable */
// @ts-nocheck

var jspb = require('google-protobuf');
var goog = jspb;
var global = (function() {
  if (this) { return this; }
  if (typeof window !== 'undefined') { return window; }
  if (typeof global !== 'undefined') { return global; }
  if (typeof self !== 'undefined') { return self; }
  return Function('return this')();
}.call(null));

goog.exportSymbol('proto.event.v1.EventType', null, global);
/**
 * @enum {number}
 */
proto.event.v1.EventType = {
  EVENT_TYPE_UNSPECIFIED: 0,
  EVENT_TYPE_USER_CREATED: 1,
  EVENT_TYPE_USER_UPDATED: 2,
  EVENT_TYPE_USER_DELETED: 3,
  EVENT_TYPE_LOGIN_SUCCEEDED: 11,
  EVENT_TYPE_LOGIN_FAILED: 12,
  EVENT_TYPE_LOGOUT: 13,
  EVENT_TYPE_TOKEN_REFRESHED: 14,
  EVENT_TYPE_TOKEN_REVOKED: 15,
  EVENT_TYPE_TENANT_TOKENS_REVOKED: 16,
  EVENT_TYPE_ROLE_CREATED: 21,
  EVENT_TYPE_ROLE_UPDATED: 22,
  EVENT_TYPE_ROLE_DELETED: 23,
  EVENT_TYPE_ROLE_ASSIGNED: 24,
  EVENT_TYPE_ROLE_REVOKED: 25,
  EVENT_TYPE_PERMISSION_CREATED: 31,
  EVENT_TYPE_PERMISSION_UPDATED: 32,
  EVENT_TYPE_PERMISSION_DELETED: 33,
  EVENT_TYPE_TENANT_CREATED: 41,
  EVENT_TYPE_TENANT_UPDATED: 42,
  EVENT_TYPE_TENANT_DELETED: 43
};

goog.object.extend(exports, proto.event.v1);
