<template>
  <el-container>
    <el-header>
      <el-row>
        <div>Mino管理面板</div>
        <el-button type="danger" icon="delete" circle @click="exit" />
      </el-row>
    </el-header>
    <el-main>
      <el-tabs v-model="activeName">
        <el-tab-pane label="配置管理" name="setting">
          <setting />
        </el-tab-pane>
        <el-tab-pane label="实时日志" name="log">
          <log />
        </el-tab-pane>
      </el-tabs>
    </el-main>
  </el-container>
</template>

<script lang="ts">
import Setting from "@/components/Setting.vue";
import Log from "@/components/Log.vue";
import { exitProgram, getSessionList } from "@/js/service";
import { defineComponent } from "vue";

export default defineComponent({
  name: "Main",
  components: {
    Setting,
    Log
  },
  data() {
    return {
      activeName: "setting",
    };
  },
  async mounted() {
    await getSessionList();
  },
  methods: {
    async exit() {
      console.log("退出程序");
      await exitProgram();
    },
  },
});
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
