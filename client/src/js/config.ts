import { mockSessionList } from "./mock";

export const IS_DEV = process.env.NODE_ENV === 'development';

export interface InterfaceConfig {
    method: string;
    path: string;
    mock?: any;
}

export const HOSTS_KEY = 'mino-host';

export const DEFAULT_HOST = (process.env.VUE_APP_HOSTS||'').length>0?process.env.VUE_APP_HOSTS:window.location.host;

export const API = {
    CONFIG_GET: {
        method: 'GET',
        path: '/api/v1/config/get',
    },
    CONFIG_SCHEMA: {
        method: 'GET',
        path: '/api/v1/config/schema',
    },
    CONFIG_SET: {
        method: 'POST',
        path: '/api/v1/config/set',
    },
    CONTROL_EXIT: {
        method: 'GET',
        path: '/api/v1/control/exit',
    },
<<<<<<< HEAD
    CONFIG_LOGIN: {
        method: 'POST',
        path: '/api/v1/login',
    },
=======
    SESSION_LIST: {
        method: 'GET',
        path: '/api/v1/session/list',
        mock: mockSessionList,
    }
>>>>>>> 9c4b06f4ff13659a6f4f9578ab54a7691ee2a29e
};

export const WS_API = {
    LOG_JSON: '/api/v1/log/json',
    LOG_TEXT: '/api/v1/log/text',
};

export const HTTP_TIMEOUT = 30000;