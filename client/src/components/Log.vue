<template>
  <div class="log-view" ref="log">
    <el-alert
      show-icon
      v-for="(item, index) in log"
      v-bind:key="index"
      :title="item.message"
      :type="item.level"
    >
    </el-alert>
  </div>
</template>

<script>
import websocket from "@/mixin/websocket";
import { getWsLogLink } from "@/js/service";

export default {
  name: "Log",
  mixins: [websocket],
  data() {
    return {
      wsLink: getWsLogLink(),
      log: [],
    };
  },
  created() {},
  mounted() {},
  methods: {
    getLevel(level) {
      switch (level) {
        case 0:
          return "error";
        case 1:
          return 'warning"';
        case 2:
          return "success";
      }
      return "info";
    },
    onWsMessage(message) {
      message.level = this.getLevel(message.level);
      this.log.push(message);
    },
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.log-view {
  max-height: 50vh;
  overflow-y: auto;
}
</style>
