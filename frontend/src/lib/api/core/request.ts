/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import axios, { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios';

import type { ApiRequestOptions } from './ApiRequestOptions';
import { CancelablePromise } from './CancelablePromise';
import type { OpenAPIConfig } from './OpenAPI';
import { getUser } from '@/lib/auth-context';

const axiosInstance = axios.create({});

export const isDefined = <T>(value: T | null | undefined): value is Exclude<T, null | undefined> => {
	return value !== undefined && value !== null;
};

export const getQueryString = (params: Record<string, any>): string => {
	const qs: string[] = [];

	const append = (key: string, value: any) => {
		qs.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`);
	};

	const process = (key: string, value: any) => {
		if (isDefined(value)) {
			if (Array.isArray(value)) {
				value.forEach(v => {
					process(key, v);
				});
			} else if (typeof value === 'object') {
				Object.entries(value).forEach(([k, v]) => {
					process(`${key}[${k}]`, v);
				});
			} else {
				append(key, value);
			}
		}
	};

	Object.entries(params).forEach(([key, value]) => {
		process(key, value);
	});

	if (qs.length > 0) {
		return `?${qs.join('&')}`;
	}

	return '';
};

const getUrl = (config: OpenAPIConfig, options: ApiRequestOptions): string => {
	const encoder = config.ENCODE_PATH || encodeURI;

	const path = options.url
		.replace('{api-version}', config.VERSION)
		.replace(/{(.*?)}/g, (substring: string, group: string) => {
			if (options.path?.hasOwnProperty(group)) {
				return encoder(String(options.path[group]));
			}
			return substring;
		});

	const url = `${config.BASE}${path}`;
	if (options.query) {
		return `${url}${getQueryString(options.query)}`;
	}
	return url;
};

export const request = <T>(
	config: OpenAPIConfig,
	options: ApiRequestOptions
): CancelablePromise<T> => {
	return new CancelablePromise((resolve, reject, onCancel) => {
		const url = getUrl(config, options);
		const token = getUser()?.token;

		const headers = {
			...options.headers,
			...(token ? { Authorization: `Basic ${token}` } : {}),
		};

		const requestConfig: AxiosRequestConfig = {
			url,
			method: options.method,
			headers,
			params: options.query ?? {},
			data: options.body ?? undefined,
		};

		const controller = new AbortController();
		requestConfig.signal = controller.signal;
		onCancel(() => controller.abort());

		axiosInstance
			.request<T>(requestConfig)
			.then((response: AxiosResponse<T>) => {
				resolve(response.data);
			})
			.catch((error: AxiosError) => {
				if (axios.isCancel(error)) {
					reject({ message: 'Request canceled', isCanceled: true });
				} else {
					reject(error);
				}
			});
	});
};
