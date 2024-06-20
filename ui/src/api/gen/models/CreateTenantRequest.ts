/* tslint:disable */
/* eslint-disable */
/**
 * api/v1/api.proto
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: version not set
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { mapValues } from '../runtime';
import type { Source } from './Source';
import {
    SourceFromJSON,
    SourceFromJSONTyped,
    SourceToJSON,
} from './Source';

/**
 * 
 * @export
 * @interface CreateTenantRequest
 */
export interface CreateTenantRequest {
    /**
     * 
     * @type {Source}
     * @memberof CreateTenantRequest
     */
    source?: Source;
}

/**
 * Check if a given object implements the CreateTenantRequest interface.
 */
export function instanceOfCreateTenantRequest(value: object): value is CreateTenantRequest {
    return true;
}

export function CreateTenantRequestFromJSON(json: any): CreateTenantRequest {
    return CreateTenantRequestFromJSONTyped(json, false);
}

export function CreateTenantRequestFromJSONTyped(json: any, ignoreDiscriminator: boolean): CreateTenantRequest {
    if (json == null) {
        return json;
    }
    return {
        
        'source': json['source'] == null ? undefined : SourceFromJSON(json['source']),
    };
}

export function CreateTenantRequestToJSON(value?: CreateTenantRequest | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'source': SourceToJSON(value['source']),
    };
}
