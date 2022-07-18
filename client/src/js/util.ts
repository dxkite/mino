import axios, { AxiosRequestConfig } from 'axios';
import { HOSTS_KEY, DEFAULT_HOST, HTTP_TIMEOUT } from './config';
import { ElNotification } from 'element-plus'

export function getApiUrl(path: string) {
    const host = window.localStorage.getItem(HOSTS_KEY) || DEFAULT_HOST;
    return "http://" + host + path
}

export function setApiHost(host: string) {
    window.localStorage.setItem(HOSTS_KEY, host)
}

export function getWsLink(path: string) {
    const host = window.localStorage.getItem(HOSTS_KEY) || DEFAULT_HOST;
    return "ws://" + host + path
}

export class ServerError {
    protected message: string;
    constructor(message: string) {
        this.message = message
    }
}


export interface Config {
    method: string;
    path: string;
}

export function requestApi(cfg: Config, data?: any) {
    const config = {
        timeout: HTTP_TIMEOUT,
    }

    let promise;
    if (cfg.method == 'POST') {
        promise = axios.post(getApiUrl(cfg.path), data, config);
    } else {
        promise = axios.get(getApiUrl(cfg.path), config);
    }
    return promise.then((data) => {
        console.log('request', config, data)
        const err = data.data.error || "";
        if (err.length > 0) {
            throw new ServerError(err);
        }
        return data.data.result;
    }).catch((e: {message: string }) => {
        console.error(e);
        ElNotification({
            type: 'error',
            title: '请求服务器失败',
            message: JSON.stringify(cfg),
        })
        if  (e.message === "need login") {
            // login
            window.location.href = '#/login';
        }
        throw e;
    })
}
