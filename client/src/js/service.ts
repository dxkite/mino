import { requestApi, getWsLink } from '@/js/util';
import { API, WS_API } from '@/js/config';


export const getConfig = () => {
  return requestApi(API.CONFIG_GET);
};

export const getConfigSchema = () => {
  return requestApi(API.CONFIG_SCHEMA).then((data) => {
    for (const name in data.properties) {
      data.properties[name]['title'] = data.properties[name].title || name;
      const uioptions: any = {
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
};

export const saveConfig = (data: any) => {
  return requestApi(API.CONFIG_SET, data).then((data) => {
    for (const name in data.properties) {
      data.properties[name]['title'] = name;
      if (data.properties[name].readOnly) {
        data.properties[name]['ui:options'] = {
          disabled: true,
          placeholder: '请输入',
        };
      }
    }
    return data;
  });
};

export const exitProgram = () => {
  return requestApi(API.CONTROL_EXIT);
};

export const getWsLogLink = () => {
  return getWsLink(WS_API.LOG_JSON);
};

export interface SessionItem {
  "id"?: number,
  "protocol"?: string,
  "group": string,
  "src"?: string,
  "dst"?: string,
  "up": number,
  "down": number,
  "closed"?: boolean,
}

export const getSessionList = async () => {
  const data = await requestApi(API.SESSION_LIST);
  console.log('处理前', data);
  const tableData = [];
  for (const [group, child] of Object.entries(data)) {
    const tableGroup: SessionItem & { children: SessionItem[] }  = { group, children: [],up: 0, down: 0};
    for (const [id, item] of Object.entries(child as SessionItem)) {
      tableGroup.children.push(item);
      tableGroup.up += item.up;
      tableGroup.down += item.down;
    }
    tableData.push(tableGroup);
  }
  console.log('处理后', tableData);
  return tableData;
};