import axios from 'axios';
import { HOSTS_KEY, DEFAULT_HOST, HTTP_TIMEOUT, InterfaceConfig, IS_DEV } from './config';
import { ElNotification } from 'element-plus';

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

const handlerError = (cfg: InterfaceConfig, e: any) => {
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
}

export const requestApi = async (cfg: InterfaceConfig, data?: any) => {
    const resp = await requestApiRaw(cfg, data);
    const err = resp.error || "";
    if (err.length > 0) {
        handlerError(cfg, new ServerError(err));
        return;
    }
    return resp.result;
}

export function requestApiRaw(cfg: InterfaceConfig, data?: any) {
    const config = {
        timeout: HTTP_TIMEOUT,
    }
    let promise;

    if (cfg.mock && IS_DEV) {
        // 模拟数据
        console.log('mock', cfg.method, cfg.path);
        promise = Promise.resolve({ data: cfg.mock });
    } else if (cfg.method == 'POST') {
        promise = axios.post(getApiUrl(cfg.path), data, config);
    } else {
        promise = axios.get(getApiUrl(cfg.path), config);
    }

    return promise.then((data) => {
        console.log('request', config, data)
        return data.data;
    }).catch((e: {message: string }) => {
        handlerError(cfg, e);
    })
}