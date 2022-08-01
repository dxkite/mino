<template>
  <div class="home">
    <header class="banner">
      <el-icon :size="14" style="margin-right: 10px">
        <Setting @click="settingVisible = true" />
      </el-icon>
      <el-icon :size="14">
        <SwitchButton @click="switchChange" />
      </el-icon>
    </header>
    <HeaderImg> </HeaderImg>
    <Table class="box-card" :tableData="tableData" @handleDisconnect="handleDisconnect" />
    <div class="journal">
      <div class="journal-header">
        <div class="spot"></div>
        <div>日志</div>
      </div>
      <div class="journal-content">
        <Log />
      </div>
    </div>
    <Footer />
  </div>
  <el-dialog v-model="settingVisible" title="设置">
    <SettingItem v-model="form" :formItem="formItem" />
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="settingVisible = false">取消</el-button>
        <el-button type="primary" @click="saveConfigChange">确定</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
import HeaderImg from "../components/HeaderImg.vue";
import SettingItem from "../components/SettingItem";
import Footer from "../components/Footer.vue";
import Table from "../components/Table.vue";
import {
  getSessionList,
  sessionClose,
  exitProgram,
  getConfigSchema,
  getConfig,
  saveConfig,
  getWsSessionLink,
} from "../js/service";
import Log from "../components/Log.vue";
import { ElMessage } from "element-plus";
import websocket from "@/mixin/websocket";

export default {
  mixins: [websocket],
  data() {
    return {
      childBorder: false,
      settingVisible: false,
      form: {},
      formItem: [],
      tableData: [], // 页面数据
      bufferData: [], // 缓存区数据
      wsLink: getWsSessionLink(),
    };
  },
  components: {
    HeaderImg,
    SettingItem,
    Log,
    Footer,
    Table,
  },
  async mounted() {
    const data = await getSessionList();
    this.bufferData = data;
    this.formItem = await getConfigSchema();

    this.form = await getConfig();
    // console.log("form", this.form);

    // 从缓冲区拿到真实数据
    this.bufferSessionData();
  },
  methods: {
    handleDisconnect(row) {
      // sessionClose();
      const data = { group: row.group, sid: row.id };
      const current = this.tableData.findIndex(
        (item) => item.src === row.group
      );
      const groupCurrent = this.tableData[current].children.findIndex(
        (item) => item.id === row.id
      );
      const groupData = this.tableData[current].children[groupCurrent];

      console.log(current, groupData);
      this.tableData[current].children.splice(groupCurrent, 1);
      if (this.tableData[current].children.length === 0) {
        this.tableData.splice(current, 1);
      }
      sessionClose(data);
    },
    switchChange() {
      console.log("switchChange");
      exitProgram();
    },
    saveConfigChange() {
      saveConfig(this.form);
      ElMessage({
        message: "保存设置成功",
        type: "success",
      });
      this.settingVisible = false;
    },
    onWsMessage(message) {
      console.log("onWsMessage", message, ", this.bufferData", this.bufferData);

      const current = this.bufferData.findIndex(
        (item) => item.src === message.info.group
      );

      // 插入不存在的id
      if (current == -1) {
        console.log("测试出不存在的id", current);
        let newTableData = {
          children: [message.info],
          down: message.info.down,
          id: message.info.group,
          isGroup: true,
          src: message.info.group,
          up: message.info.up,
        };
        // 插入缓冲区
        this.bufferData.push(newTableData);
        return;
      }

      const groupCurrent = this.bufferData[current].children.findIndex(
        (item) => item.id === message.info.id
      );

      if (groupCurrent == -1) {
        // 插入缓冲区
        this.bufferData[current].children.push(message.info);
        return;
      } else {
        // 更新数据
        this.bufferData[current].children[groupCurrent] = message.info;
      }

      // 父组件流量求和
      const upSum = this.bufferData[current].children.reduce((prev, item) => {
        return prev + item.up;
      }, 0);
      this.bufferData[current].up = upSum;
      const downSum = this.bufferData[current].children.reduce((prev, item) => {
        return prev + item.down;
      }, 0);
      this.bufferData[current].down = downSum;

      // 删除更新
      if (message.type == "close") {
        console.log("删除数据");
        this.bufferData[current].children.splice(groupCurrent, 1);
        if (this.bufferData[current].children.length === 0) {
          this.bufferData.splice(current, 1);
        }
      }
    },
    // 暂存会话数据
    bufferSessionData() {
      setInterval(() => {
        this.tableData = this.bufferData;
        console.log("*tableData", this.tableData);
      }, 1000);
    },
  },
};
</script>

<style scoped>
.home {
  display: flex;
  align-items: center;
  flex-direction: column;
}

.banner {
  display: flex;
  margin-top: 30px;
  justify-content: flex-end;
  height: 26px;
  width: 80%;
}

.box-card {
  margin-top: 53px;
  width: 80%;
}

.journal {
  box-sizing: border-box;
  margin-top: 64px;
  width: 80%;
  padding: 15px;
  border: 1px solid #e1f3d8;
}

.journal-header {
  display: flex;
  align-items: center;
  padding: 4px;
}

.spot {
  margin-right: 7px;
}

.red {
  color: #c45656;
}

.spot {
  height: 8px;
  width: 8px;
  border-radius: 8px;
  background: #95d475;
}
</style>
