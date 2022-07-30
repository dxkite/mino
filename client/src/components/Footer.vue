<template>
  <div class="footer">mino 网络访问助手 {{ version }} {{ commit }}</div>
</template>

<script>
import { getStatus, } from '@/js/service'
export default {
  name: "footer",
  data() {
    return {
      version: "",
      commit: "",
    };
  },
  async mounted() {
    // 版本信息
    const { version, commit } = await getStatus();
    this.version = version;
    this.commit = commit;
    console.log("getStatus", version, commit);
  },
};
</script>

<style scoped>
.footer {
  padding: 40px 0;
  font-family: "Inter";
  font-style: normal;
  font-weight: 400;
  font-size: 14px;
  line-height: 17px;
}
</style>