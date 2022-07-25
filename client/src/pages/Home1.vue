<template>
  <div class="home">
    <header class="header-icon">
      <el-icon :size="14" style="margin-right: 10px">
        <Setting @click="settingVisible = true" />
      </el-icon>
      <el-icon :size="14">
        <SwitchButton @click="switchChange" />
      </el-icon>
    </header>
    <HeaderImg> </HeaderImg>
    <div class="box-card">
      <el-table :data="tableData" style="width: 100%" row-key="id" border>
        <el-table-column label="远程客户端" prop="src" width="180" />
        <el-table-column label="协议" prop="protocol" width="80" />
        <el-table-column label="目标网站" prop="dst">
          <template #default="scope">
            <div :title="scope.row.dst" class="table-dst">
              {{ scope.row.dst }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="上传流量" prop="up" width="100" />
        <el-table-column label="下载流量" prop="down" width="100" />
        <el-table-column label="操作" prop="state" width="100">
          <template #default="scope">
            <el-button
              v-if="!scope.row.isGroup"
              size="small"
              type="danger"
              @click="handleDisconnect(scope.row)"
              >断开连接</el-button
            >
          </template>
        </el-table-column>
      </el-table>
    </div>
    <div class="journal">
      <div class="journal-header">
        <div class="spot"></div>
        <div>日志</div>
      </div>
      <div class="journal-cotent">
        <Log />
      </div>
    </div>
    <div class="footer">mino 网络访问助手 v0.2.6-beta a82db2</div>
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
import {
  getSessionList,
  sessionClose,
  exitProgram,
  getConfigSchema,
  getConfig,
  saveConfig,
} from "../js/service";
import Log from "../components/Log.vue";
import { ElMessage } from 'element-plus'

export default {
  data() {
    return {
      childBorder: false,
      settingVisible: false,
      form: {},
      formItem: [],
      tableData: [],
    };
  },
  components: {
    HeaderImg,
    SettingItem,
    Log,
  },
  async mounted() {
    const data = await getSessionList();
    this.tableData = data;
    // console.log("tabaleData", data);
    this.formItem = await getConfigSchema();
    // console.log('getConfigSchema',this.form)

    this.form = await getConfig();
    console.log("form", this.form);
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
  },
};
</script>

<style scoped>
.home {
  display: flex;
  align-items: center;
  flex-direction: column;
}
.header-icon {
  display: flex;
  margin-top: 30px;
  justify-content: flex-end;
  height: 26px;
  width: 1052px;
}
.box-card {
  margin-top: 53px;
  width: 1052px;
}
.journal {
  box-sizing: border-box;
  margin-top: 64px;
  width: 1052px;
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
.footer {
  margin-top: 32px;
  font-family: "Inter";
  font-style: normal;
  font-weight: 400;
  font-size: 14px;
  line-height: 17px;
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
.table-dst {
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
}
</style>
