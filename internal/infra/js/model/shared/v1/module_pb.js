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
var global =
    (typeof globalThis !== 'undefined' && globalThis) ||
    (typeof window !== 'undefined' && window) ||
    (typeof global !== 'undefined' && global) ||
    (typeof self !== 'undefined' && self) ||
    (function () { return this; }).call(null) ||
    Function('return this')();

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
