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


import * as runtime from '../runtime';
import type {
  CreateTenantRequest,
  CreateTenantResponse,
  DeleteTenantResponse,
  GetTenantResponse,
  ListTenantsResponse,
  RpcStatus,
  TenantServiceUpdateTenantBody,
  UpdateTenantResponse,
} from '../models/index';
import {
    CreateTenantRequestFromJSON,
    CreateTenantRequestToJSON,
    CreateTenantResponseFromJSON,
    CreateTenantResponseToJSON,
    DeleteTenantResponseFromJSON,
    DeleteTenantResponseToJSON,
    GetTenantResponseFromJSON,
    GetTenantResponseToJSON,
    ListTenantsResponseFromJSON,
    ListTenantsResponseToJSON,
    RpcStatusFromJSON,
    RpcStatusToJSON,
    TenantServiceUpdateTenantBodyFromJSON,
    TenantServiceUpdateTenantBodyToJSON,
    UpdateTenantResponseFromJSON,
    UpdateTenantResponseToJSON,
} from '../models/index';

export interface TenantServiceCreateTenantRequest {
    body: CreateTenantRequest;
}

export interface TenantServiceDeleteTenantRequest {
    id: string;
}

export interface TenantServiceGetTenantRequest {
    id: string;
}

export interface TenantServiceUpdateTenantRequest {
    id: string;
    body: TenantServiceUpdateTenantBody;
}

/**
 * 
 */
export class TenantServiceApi extends runtime.BaseAPI {

    /**
     */
    async tenantServiceCreateTenantRaw(requestParameters: TenantServiceCreateTenantRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<CreateTenantResponse>> {
        if (requestParameters['body'] == null) {
            throw new runtime.RequiredError(
                'body',
                'Required parameter "body" was null or undefined when calling tenantServiceCreateTenant().'
            );
        }

        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/json';

        const response = await this.request({
            path: `/v1/tenants`,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: CreateTenantRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => CreateTenantResponseFromJSON(jsonValue));
    }

    /**
     */
    async tenantServiceCreateTenant(requestParameters: TenantServiceCreateTenantRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<CreateTenantResponse> {
        const response = await this.tenantServiceCreateTenantRaw(requestParameters, initOverrides);
        return await response.value();
    }

    /**
     */
    async tenantServiceDeleteTenantRaw(requestParameters: TenantServiceDeleteTenantRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<DeleteTenantResponse>> {
        if (requestParameters['id'] == null) {
            throw new runtime.RequiredError(
                'id',
                'Required parameter "id" was null or undefined when calling tenantServiceDeleteTenant().'
            );
        }

        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        const response = await this.request({
            path: `/v1/tenants/{id}`.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters['id']))),
            method: 'DELETE',
            headers: headerParameters,
            query: queryParameters,
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => DeleteTenantResponseFromJSON(jsonValue));
    }

    /**
     */
    async tenantServiceDeleteTenant(requestParameters: TenantServiceDeleteTenantRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<DeleteTenantResponse> {
        const response = await this.tenantServiceDeleteTenantRaw(requestParameters, initOverrides);
        return await response.value();
    }

    /**
     */
    async tenantServiceGetTenantRaw(requestParameters: TenantServiceGetTenantRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<GetTenantResponse>> {
        if (requestParameters['id'] == null) {
            throw new runtime.RequiredError(
                'id',
                'Required parameter "id" was null or undefined when calling tenantServiceGetTenant().'
            );
        }

        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        const response = await this.request({
            path: `/v1/tenants/{id}`.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters['id']))),
            method: 'GET',
            headers: headerParameters,
            query: queryParameters,
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => GetTenantResponseFromJSON(jsonValue));
    }

    /**
     */
    async tenantServiceGetTenant(requestParameters: TenantServiceGetTenantRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<GetTenantResponse> {
        const response = await this.tenantServiceGetTenantRaw(requestParameters, initOverrides);
        return await response.value();
    }

    /**
     */
    async tenantServiceListTenantsRaw(initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<ListTenantsResponse>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        const response = await this.request({
            path: `/v1/tenants`,
            method: 'GET',
            headers: headerParameters,
            query: queryParameters,
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => ListTenantsResponseFromJSON(jsonValue));
    }

    /**
     */
    async tenantServiceListTenants(initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<ListTenantsResponse> {
        const response = await this.tenantServiceListTenantsRaw(initOverrides);
        return await response.value();
    }

    /**
     */
    async tenantServiceUpdateTenantRaw(requestParameters: TenantServiceUpdateTenantRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<UpdateTenantResponse>> {
        if (requestParameters['id'] == null) {
            throw new runtime.RequiredError(
                'id',
                'Required parameter "id" was null or undefined when calling tenantServiceUpdateTenant().'
            );
        }

        if (requestParameters['body'] == null) {
            throw new runtime.RequiredError(
                'body',
                'Required parameter "body" was null or undefined when calling tenantServiceUpdateTenant().'
            );
        }

        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/json';

        const response = await this.request({
            path: `/v1/tenants/{id}`.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters['id']))),
            method: 'PUT',
            headers: headerParameters,
            query: queryParameters,
            body: TenantServiceUpdateTenantBodyToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => UpdateTenantResponseFromJSON(jsonValue));
    }

    /**
     */
    async tenantServiceUpdateTenant(requestParameters: TenantServiceUpdateTenantRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<UpdateTenantResponse> {
        const response = await this.tenantServiceUpdateTenantRaw(requestParameters, initOverrides);
        return await response.value();
    }

}
