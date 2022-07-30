import { requestApi, getWsLink, requestApiRaw } from '@/js/util';
import { API, WS_API } from '@/js/config';


export const getConfig = () => {
  return requestApi(API.CONFIG_GET);
};

// export const getConfigSchema = () => {
//   return requestApi(API.CONFIG_SCHEMA).then((data) => {
//     for (const name in data.properties) {
//       data.properties[name]['title'] = data.properties[name].title || name;
//       const uioptions: any = {
//         placeholder: '请输入'
//       };
//       if (data.properties[name].readOnly) {
//         uioptions['disabled'] = true;
//       }
//       if (data.properties[name].description) {
//         uioptions['description'] = data.properties[name].description;
//       }
//       data.properties[name]['ui:options'] = uioptions;
//     }
//     return data;
//   })
// };
export const getConfigSchema = () => {
  return requestApi(API.CONFIG_SCHEMA).then((data) => {
    for (const name in data.properties) {
      data.properties[name]['title'] = data.properties[name].title || name;
      data.properties[name]['name'] = name;
    }
    return Object.values(data.properties);
  })
};

// export const saveConfig = (data: any) => {
//   return requestApi(API.CONFIG_SET, data).then((data) => {
//     for (const name in data.properties) {
//       data.properties[name]['title'] = name;
//       if (data.properties[name].readOnly) {
//         data.properties[name]['ui:options'] = {
//           disabled: true,
//           placeholder: '请输入',
//         };
//       }
//     }
//     return data;
//   });
// };
export const saveConfig = (data: any) => {
  return requestApi(API.CONFIG_SET, data).then((data) => {
    return data;
  });
}

export const exitProgram = () => {
  return requestApi(API.CONTROL_EXIT);
};

export const getWsLogLink = () => {
  return getWsLink(WS_API.LOG_JSON);
};

export const getWsSessionLink = () => {
  return getWsLink(WS_API.SESSION);
};

export interface SessionItem {
  "id"?: number|string,
  "protocol"?: string,
  "group"?: string,
  "src": string,
  "dst"?: string,
  "up": number,
  "down": number,
  "closed"?: boolean,
  "isGroup"?: boolean,
}

export const getSessionList = async () => {
  const data = await requestApi(API.SESSION_LIST);
  console.log('处理前', data);
  const tableData = [];
  for (const [group, child] of Object.entries(data)) {
    const tableGroup: SessionItem & { children: SessionItem[] }  = { src: group, children: [],up: 0, down: 0, id: group};
    for (const [id, item] of Object.entries(child as SessionItem)) {
      tableGroup.children.push(item);
      tableGroup.up += item.up;
      tableGroup.down += item.down;
      tableGroup.isGroup = true;
    }
    if (tableGroup.children.length > 0) {
      tableData.push(tableGroup);
    }
  }
  console.log('处理后', tableData);
  return tableData;
};

export const sessionClose = (data: {group: string, sid: number}) => {
  return requestApi(API.SESSION_CLOSE, data);
}

export const login = (data: {username: string, password: string}) => {
  return requestApi(API.CONFIG_LOGIN, data)
};

export const getStatus = () =>  {
  return requestApiRaw(API.STATUS);
}