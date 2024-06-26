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
import type { Application } from './Application';
import {
    ApplicationFromJSON,
    ApplicationFromJSONTyped,
    ApplicationToJSON,
} from './Application';
import type { Source } from './Source';
import {
    SourceFromJSON,
    SourceFromJSONTyped,
    SourceToJSON,
} from './Source';

/**
 * 
 * @export
 * @interface Tenant
 */
export interface Tenant {
    /**
     * 
     * @type {string}
     * @memberof Tenant
     */
    id?: string;
    /**
     * 
     * @type {Source}
     * @memberof Tenant
     */
    source?: Source;
    /**
     * 
     * @type {Application}
     * @memberof Tenant
     */
    application?: Application;
}

/**
 * Check if a given object implements the Tenant interface.
 */
export function instanceOfTenant(value: object): value is Tenant {
    return true;
}

export function TenantFromJSON(json: any): Tenant {
    return TenantFromJSONTyped(json, false);
}

export function TenantFromJSONTyped(json: any, ignoreDiscriminator: boolean): Tenant {
    if (json == null) {
        return json;
    }
    return {
        
        'id': json['id'] == null ? undefined : json['id'],
        'source': json['source'] == null ? undefined : SourceFromJSON(json['source']),
        'application': json['application'] == null ? undefined : ApplicationFromJSON(json['application']),
    };
}

export function TenantToJSON(value?: Tenant | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'id': value['id'],
        'source': SourceToJSON(value['source']),
        'application': ApplicationToJSON(value['application']),
    };
}

