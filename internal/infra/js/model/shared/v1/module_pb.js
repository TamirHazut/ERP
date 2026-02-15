// source: shared/v1/module.proto
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

goog.exportSymbol('proto.shared.v1.Module', null, global);
/**
 * @enum {number}
 */
proto.shared.v1.Module = {
  MODULE_UNSPECIFIED: 0,
  MODULE_AUTH: 1,
  MODULE_CONFIG: 2,
  MODULE_CORE: 3,
  MODULE_DB: 4,
  MODULE_EVENT: 5,
  MODULE_GATEWAY: 6,
  MODULE_INIT: 7,
  MODULE_SIDECAR: 8,
  MODULE_WEBUI: 9
};

goog.object.extend(exports, proto.shared.v1);
