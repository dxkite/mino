(function(e){function t(t){for(var o,c,a=t[0],s=t[1],u=t[2],f=0,d=[];f<a.length;f++)c=a[f],Object.prototype.hasOwnProperty.call(r,c)&&r[c]&&d.push(r[c][0]),r[c]=0;for(o in s)Object.prototype.hasOwnProperty.call(s,o)&&(e[o]=s[o]);l&&l(t);while(d.length)d.shift()();return i.push.apply(i,u||[]),n()}function n(){for(var e,t=0;t<i.length;t++){for(var n=i[t],o=!0,a=1;a<n.length;a++){var s=n[a];0!==r[s]&&(o=!1)}o&&(i.splice(t--,1),e=c(c.s=n[0]))}return e}var o={},r={app:0},i=[];function c(t){if(o[t])return o[t].exports;var n=o[t]={i:t,l:!1,exports:{}};return e[t].call(n.exports,n,n.exports,c),n.l=!0,n.exports}c.m=e,c.c=o,c.d=function(e,t,n){c.o(e,t)||Object.defineProperty(e,t,{enumerable:!0,get:n})},c.r=function(e){"undefined"!==typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(e,"__esModule",{value:!0})},c.t=function(e,t){if(1&t&&(e=c(e)),8&t)return e;if(4&t&&"object"===typeof e&&e&&e.__esModule)return e;var n=Object.create(null);if(c.r(n),Object.defineProperty(n,"default",{enumerable:!0,value:e}),2&t&&"string"!=typeof e)for(var o in e)c.d(n,o,function(t){return e[t]}.bind(null,o));return n},c.n=function(e){var t=e&&e.__esModule?function(){return e["default"]}:function(){return e};return c.d(t,"a",t),t},c.o=function(e,t){return Object.prototype.hasOwnProperty.call(e,t)},c.p="/";var a=window["webpackJsonp"]=window["webpackJsonp"]||[],s=a.push.bind(a);a.push=t,a=a.slice();for(var u=0;u<a.length;u++)t(a[u]);var l=s;i.push([0,"chunk-vendors"]),n()})({0:function(e,t,n){e.exports=n("56d7")},"419c":function(e,t,n){},"43f7":function(e,t,n){},"56d7":function(e,t,n){"use strict";n.r(t);n("e260"),n("e6cf"),n("cca6"),n("a79d");var o=n("7a23");function r(e,t,n,r,i,c){var a=Object(o["O"])("Main");return Object(o["F"])(),Object(o["k"])(a)}var i=Object(o["ib"])("data-v-5b557a4c");Object(o["I"])("data-v-5b557a4c");var c=Object(o["n"])(" Mino管理面板 "),a=Object(o["n"])(" 关闭程序 ");Object(o["G"])();var s=i((function(e,t,n,r,s,u){var l=Object(o["O"])("el-button"),f=Object(o["O"])("el-header"),d=Object(o["O"])("setting"),p=Object(o["O"])("el-tab-pane"),h=Object(o["O"])("log"),g=Object(o["O"])("el-tabs"),O=Object(o["O"])("el-main"),b=Object(o["O"])("el-container");return Object(o["F"])(),Object(o["k"])(b,null,{default:i((function(){return[Object(o["o"])(f,null,{default:i((function(){return[c,Object(o["o"])(l,{size:"mini",type:"danger",onClick:u.exit},{default:i((function(){return[a]})),_:1},8,["onClick"])]})),_:1}),Object(o["o"])(O,null,{default:i((function(){return[Object(o["o"])(g,{modelValue:s.activeName,"onUpdate:modelValue":t[1]||(t[1]=function(e){return s.activeName=e})},{default:i((function(){return[Object(o["o"])(p,{label:"配置管理",name:"setting"},{default:i((function(){return[Object(o["o"])(d)]})),_:1}),Object(o["o"])(p,{label:"实时日志",name:"log"},{default:i((function(){return[Object(o["o"])(h)]})),_:1})]})),_:1},8,["modelValue"])]})),_:1})]})),_:1})})),u=Object(o["ib"])("data-v-0a964488"),l=u((function(e,t,n,r,i,c){var a=Object(o["O"])("vue-form");return Object(o["F"])(),Object(o["k"])(a,{modelValue:i.formData,"onUpdate:modelValue":t[1]||(t[1]=function(e){return i.formData=e}),"ui-schema":e.uiSchema,schema:i.schema,formProps:i.formProps,onSubmit:c.submit,onCancel:c.cancel},null,8,["modelValue","ui-schema","schema","formProps","onSubmit","onCancel"])})),f=n("d4ec"),d=n("bc3a"),p=n.n(d),h="mino-host",g=window.location.host,O={CONFIG_GET:{method:"GET",path:"/api/v1/config/get"},CONFIG_SCHEMA:{method:"GET",path:"/api/v1/config/schema"},CONFIG_SET:{method:"POST",path:"/api/v1/config/set"},CONTROL_EXIT:{method:"GET",path:"/api/v1/control/exit"}},b={LOG_JSON:"/api/v1/log/json",LOG_TEXT:"/api/v1/log/text"},m=3e4,v=n("7864");function j(e){var t=window.localStorage.getItem(h)||g;return"http://"+t+e}function w(e){var t=window.localStorage.getItem(h)||g;return"ws://"+t+e}var _=function e(t){Object(f["a"])(this,e),this.message=t};function y(e,t){var n,o={timeout:m};return n="POST"==e.method?p.a.post(j(e.path),t,o):p.a.get(j(e.path),o),n.then((function(e){console.log("request",o,e);var t=e.data.error||"";if(t.length>0)throw new _(t);return e.data.result})).catch((function(t){throw console.error(t),Object(v["a"])({type:"error",title:"请求服务器失败",message:JSON.stringify(e)}),t}))}var k={getConfig:function(){return y(O.CONFIG_GET)},getConfigSchema:function(){return y(O.CONFIG_SCHEMA).then((function(e){for(var t in e.properties)e.properties[t]["title"]=t,e.properties[t].readOnly&&(e.properties[t]["ui:options"]={disabled:!0,placeholder:"请输入"});return e}))},saveConfig:function(e){return y(O.CONFIG_SET,e).then((function(e){for(var t in e.properties)e.properties[t]["title"]=t,e.properties[t].readOnly&&(e.properties[t]["ui:options"]={disabled:!0,placeholder:"请输入"});return e}))},exitProgram:function(){return y(O.CONTROL_EXIT)},getWsLogLink:function(){return w(b.LOG_JSON)}},C={name:"Setting",data:function(){return{formData:{},defaultData:{},schema:{},formProps:{}}},created:function(){var e=this;k.getConfigSchema().then((function(t){console.log("schema",t),e.schema=t}))},mounted:function(){console.log("mounted"),this.getConfig()},methods:{submit:function(){var e=this;console.log("submit"),k.saveConfig(this.formData).then((function(){e.$notify({type:"success",message:"配置更新成功"}),e.getConfig()}))},cancel:function(){console.log("cancel"),this.formData=this.defaultData},getConfig:function(){var e=this;k.getConfig().then((function(t){console.log("data",t),e.formData=t,e.defaultData=t}))}}};C.render=l,C.__scopeId="data-v-0a964488";var S=C,L=Object(o["ib"])("data-v-099c1d85");Object(o["I"])("data-v-099c1d85");var T={class:"log-view",ref:"log"};Object(o["G"])();var E=L((function(e,t,n,r,i,c){var a=Object(o["O"])("el-alert");return Object(o["F"])(),Object(o["k"])("div",T,[(Object(o["F"])(!0),Object(o["k"])(o["b"],null,Object(o["M"])(i.log,(function(e,t){return Object(o["F"])(),Object(o["k"])(a,{"show-icon":"",key:t,title:e.message,type:e.level},null,8,["title","type"])})),128))],512)})),G={data:function(){return{wsLink:"",wsType:"json",socket:null}},mounted:function(){this._ws$init()},methods:{_ws$init:function(){var e=this;console.log("Init Websocket"),this.socket=new WebSocket(this.wsLink),this.socket.addEventListener("open",(function(){console.log("open"),e._ws$open()})),this.socket.addEventListener("close",(function(t){console.log("close",t),e._ws$close()})),this.socket.addEventListener("error",(function(t){console.log("error",t),e._ws$error()})),this.socket.addEventListener("message",(function(t){e._ws$message(t.data)}))},_ws$open:function(){console.log("open"),this.onWsOpen&&this.onWsOpen()},_ws$message:function(e){this.onWsMessage&&("json"===this.wsType&&(e=JSON.parse(e)),this.onWsMessage(e))},$send:function(e){this.socket.send(e)},_ws$close:function(){this.onWsClose&&this.onWsClose()},_ws$error:function(){this.onWsError?this.onWsError():this.$notify({type:"error",title:"链接websocket失败",message:this.wsLink})}}},I=G,N={name:"Log",mixins:[I],data:function(){return{wsLink:k.getWsLogLink(),log:[]}},created:function(){},mounted:function(){},methods:{getLevel:function(e){switch(e){case 0:return"error";case 1:return'warning"';case 2:return"success"}return"info"},onWsMessage:function(e){e.level=this.getLevel(e.level),this.log.push(e)}}};n("efc9");N.render=E,N.__scopeId="data-v-099c1d85";var P=N,$={name:"Main",components:{Setting:S,Log:P},data:function(){return{activeName:"setting"}},methods:{exit:function(){var e=this;console.log("退出程序"),k.exitProgram().then((function(){e.$notify({type:"error",message:"关闭失败"})})).catch((function(){e.$notify({type:"success",message:"关闭成功"})}))}}};$.render=s,$.__scopeId="data-v-5b557a4c";var M=$,x={name:"App",components:{Main:M}};n("952f");x.render=r;var W=x,F=(n("7dd6"),n("3ef0")),D=n.n(F),J=function(e){e.use(v["b"],{locale:D.a})},V=n("d6f7"),A=Object(o["j"])(W);J(A),A.use(V["a"]),A.mount("#app")},"952f":function(e,t,n){"use strict";n("43f7")},efc9:function(e,t,n){"use strict";n("419c")}});
//# sourceMappingURL=app.565afeb7.js.map