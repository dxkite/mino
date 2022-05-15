import { requestApi, getWsLink } from '@/js/util';
import { API, WS_API } from '@/js/config';

export default {
    getConfig() {
        return requestApi(API.CONFIG_GET)
    },
    getConfigSchema() {
        return requestApi(API.CONFIG_SCHEMA).then((data) => {
            for (let name in data.properties) {
                data.properties[name]['title'] = data.properties[name].title || name;
                let uioptions = {
                    placeholder: '请输入'
                };
                if (data.properties[name].readOnly) {
                    uioptions['disabled'] = true;
                }
                if (data.properties[name].description) {
                    uioptions['description'] = data.properties[name].description;
                }
                data.properties[name]['ui:options'] = uioptions;
            }
            return data;
        })
    },
    saveConfig(data) {
        return requestApi(API.CONFIG_SET, data).then((data) => {
            for (let name in data.properties) {
                data.properties[name]['title'] = name;
                if (data.properties[name].readOnly) {
                    data.properties[name]['ui:options'] = {
                        disabled: true,
                        placeholder: '请输入',
                    };
                }
            }
            return data;
        })
    },
    exitProgram() {
        return requestApi(API.CONTROL_EXIT)
    },
    getWsLogLink() {
        return getWsLink(WS_API.LOG_JSON)
    },
}