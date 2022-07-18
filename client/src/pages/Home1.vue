<template>
  <div class="home">
    <header class="header-icon">
      <el-icon :size="14">
        <Setting @click="settingVisible = true" />
      </el-icon>
      <el-icon :size="14">
        <SwitchButton @click="switchChange" />
      </el-icon>
    </header>
    <HeaderImg> </HeaderImg>
    <div class="box-card">
      <el-table
        :data="tableData"
        style="width: 100%"
        row-key="id"
        border
        lazy
        :load="load"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      >
        <el-table-column label="远程客户端" prop="src" />
        <el-table-column label="协议" prop="protocol" />
        <el-table-column label="目标网站" prop="dst" />
        <el-table-column label="上传流量" prop="up" />
        <el-table-column label="下载流量" prop="down" />
        <el-table-column label="操作" prop="state">
          <template #default="scope">
            <el-button
              size="small"
              type="danger"
              @click="handleDisconnect(scope.$index, scope.row)"
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
        <div :class="stylerColor">12121221</div>
      </div>
    </div>
    <div class="footer">mino 网络访问助手 v0.2.6-beta a82db2</div>
  </div>
  <el-dialog v-model="settingVisible" title="设置">
    <SettingItem v-model="form" />
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="settingVisible = false">取消</el-button>
        <el-button type="primary" @click="settingVisible = false"
          >确定</el-button
        >
      </span>
    </template>
  </el-dialog>
</template>

<script>
import HeaderImg from "../components/HeaderImg.vue";
import SettingItem from "../components/SettingItem";
import { getSessionList } from '../js/service';

export default {
  data() {
    return {
      stylerColor: "green",
      childBorder: false,
      settingVisible: false,
      form: {
        address: "",
        host_detect_loopback: false,
        hot_load: 10000,
      },
      tableData: [],
    };
  },
  components: {
    HeaderImg,
    SettingItem,
  },
  async mounted() {
    const data = await getSessionList();
    this.tableData = data;
    console.log(data);
  },
  methods: {
    handleDisconnect(e) {
      console.log("handleDisconnect", e);
    },
    switchChange() {
      console.log("switchChange");
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
  height: 251px;
  width: 1052px;
  padding: 15px;
  border: 1px solid #e1f3d8;
}
.journal-header {
  display: flex;
  align-items: center;
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
.green {
  color: #95d475;
}
.spot {
  height: 8px;
  width: 8px;
  border-radius: 8px;
  background: #95d475;
}
</style>
