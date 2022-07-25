<template>
  <div class="log-view" ref="log">
    <div
      :class="item2.level"
      class="content"
      v-for="(item2, index) in log"
      :key="index"
      :title="item2.message"
    >
      {{ item2.time }}
      [{{ item2.level.toUpperCase() }}]
      {{ item2.message }}
    </div>
    <!-- <el-alert
      show-icon
      v-for="(item, index) in log"
      v-bind:key="index"
      :title="item.message"
      :type="item.level"
    >
    </el-alert> -->
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
      stylerColor: "success",
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
    getTime(date) {
      var json_date = new Date(date).toJSON();
      return new Date(new Date(json_date) + 8 * 3600 * 1000)
        .toISOString()
        .replace(/T/g, " ")
        .replace(/\.[\d]{3}Z/, "");
    },
    onWsMessage(message) {
      // console.log(message);
      message.level = this.getLevel(message.level);
      message.time = this.getTime(message.time);
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
  font-family: Roboto;
  font-size: 14px;
  letter-spacing: 0px;
  text-align: left;
}

::-webkit-scrollbar {
  width: 6px;
}

::-webkit-scrollbar-thumb {
  /* 滚动条样式 */
  width: 6px;
  height: 113px;

  background: #d9d9d9;
  border-radius: 4px;
}
.success {
  color: #95d475;
}
.error {
  color: #c45656;
}
.info {
  color: #000000cc;
}
.content {
  margin-bottom: 4px;

  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
}
</style>
