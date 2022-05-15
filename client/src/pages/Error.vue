<template>
  <div class="error">
    <div class="banner">
      <div class="icon-setting" @click="onClickSetting">
        <svg
          viewBox="0 0 1024 1024"
          xmlns="http://www.w3.org/2000/svg"
          data-v-ba633cb8=""
        >
          <path
            fill="currentColor"
            d="M600.704 64a32 32 0 0 1 30.464 22.208l35.2 109.376c14.784 7.232 28.928 15.36 42.432 24.512l112.384-24.192a32 32 0 0 1 34.432 15.36L944.32 364.8a32 32 0 0 1-4.032 37.504l-77.12 85.12a357.12 357.12 0 0 1 0 49.024l77.12 85.248a32 32 0 0 1 4.032 37.504l-88.704 153.6a32 32 0 0 1-34.432 15.296L708.8 803.904c-13.44 9.088-27.648 17.28-42.368 24.512l-35.264 109.376A32 32 0 0 1 600.704 960H423.296a32 32 0 0 1-30.464-22.208L357.696 828.48a351.616 351.616 0 0 1-42.56-24.64l-112.32 24.256a32 32 0 0 1-34.432-15.36L79.68 659.2a32 32 0 0 1 4.032-37.504l77.12-85.248a357.12 357.12 0 0 1 0-48.896l-77.12-85.248A32 32 0 0 1 79.68 364.8l88.704-153.6a32 32 0 0 1 34.432-15.296l112.32 24.256c13.568-9.152 27.776-17.408 42.56-24.64l35.2-109.312A32 32 0 0 1 423.232 64H600.64zm-23.424 64H446.72l-36.352 113.088-24.512 11.968a294.113 294.113 0 0 0-34.816 20.096l-22.656 15.36-116.224-25.088-65.28 113.152 79.68 88.192-1.92 27.136a293.12 293.12 0 0 0 0 40.192l1.92 27.136-79.808 88.192 65.344 113.152 116.224-25.024 22.656 15.296a294.113 294.113 0 0 0 34.816 20.096l24.512 11.968L446.72 896h130.688l36.48-113.152 24.448-11.904a288.282 288.282 0 0 0 34.752-20.096l22.592-15.296 116.288 25.024 65.28-113.152-79.744-88.192 1.92-27.136a293.12 293.12 0 0 0 0-40.256l-1.92-27.136 79.808-88.128-65.344-113.152-116.288 24.96-22.592-15.232a287.616 287.616 0 0 0-34.752-20.096l-24.448-11.904L577.344 128zM512 320a192 192 0 1 1 0 384 192 192 0 0 1 0-384zm0 64a128 128 0 1 0 0 256 128 128 0 0 0 0-256z"
          ></path>
        </svg>
      </div>
    </div>
    <div class="message">
      <el-row justify="center">
        <div class="action-name">{{ actionName }}</div>
        <div class="detail">
          <div class="domain">网站 {{ domain }}</div>
          <div v-if="isError" class="error-info">
            <div class="error-message">网站连接失败，请尝试以下操作</div>
            <el-tooltip
              class="box-item"
              effect="dark"
              :content="error"
              placement="top"
            >
              <div class="icon-question">
                <svg
                  viewBox="0 0 1024 1024"
                  xmlns="http://www.w3.org/2000/svg"
                  data-v-ba633cb8=""
                >
                  <path
                    fill="currentColor"
                    d="M512 64a448 448 0 1 1 0 896 448 448 0 0 1 0-896zm23.744 191.488c-52.096 0-92.928 14.784-123.2 44.352-30.976 29.568-45.76 70.4-45.76 122.496h80.256c0-29.568 5.632-52.8 17.6-68.992 13.376-19.712 35.2-28.864 66.176-28.864 23.936 0 42.944 6.336 56.32 19.712 12.672 13.376 19.712 31.68 19.712 54.912 0 17.6-6.336 34.496-19.008 49.984l-8.448 9.856c-45.76 40.832-73.216 70.4-82.368 89.408-9.856 19.008-14.08 42.24-14.08 68.992v9.856h80.96v-9.856c0-16.896 3.52-31.68 10.56-45.76 6.336-12.672 15.488-24.64 28.16-35.2 33.792-29.568 54.208-48.576 60.544-55.616 16.896-22.528 26.048-51.392 26.048-86.592 0-42.944-14.08-76.736-42.24-101.376-28.16-25.344-65.472-37.312-111.232-37.312zm-12.672 406.208a54.272 54.272 0 0 0-38.72 14.784 49.408 49.408 0 0 0-15.488 38.016c0 15.488 4.928 28.16 15.488 38.016A54.848 54.848 0 0 0 523.072 768c15.488 0 28.16-4.928 38.72-14.784a51.52 51.52 0 0 0 16.192-38.72 51.968 51.968 0 0 0-15.488-38.016 55.936 55.936 0 0 0-39.424-14.784z"
                  ></path>
                </svg>
              </div>
            </el-tooltip>
          </div>
          <div v-else>根据网络访问配置，该网站禁止访问</div>
        </div>
      </el-row>
      <el-row v-if="isError" justify="center" class="action-panel">
        <el-button @click="onClickRetry">重试访问</el-button>
      </el-row>
    </div>
    <div class="footer">mino 网络访问助手 {{ version }} {{ commitHash }}</div>
  </div>
</template>


<script>
export default {
  name: "Error",
  components: {},

  data() {
    return {
      action: "error",
      domain: "",
      error: "",
      url: "",
      version: "",
      commitHash: "",
    };
  },
  computed: {
    isError() {
      return this.action === "error";
    },
    actionName() {
      return this.isError ? "连接失败" : "禁止访问";
    },
  },
  mounted() {
    const data = JSON.parse(this.$route.query.action || "{}");
    this.action = data.action;
    this.domain = data.domain;
    this.error = data.error;
    this.url = data.url;
    this.version = data.version;
    this.commitHash = data.commit;
  },
  methods: {
    onClickSetting() {
      this.$router.push({ name: "Home" });
    },
    onClickRetry() {
      window.location.href = this.url;
    },
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.error {
  height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.banner {
  padding: 1em;
  display: flex;
  justify-content: flex-end;
}

.icon-setting {
  width: 32px;
  height: 32px;
}

.icon-setting:hover {
  width: 32px;
  height: 32px;
  color: #409eff;
}

.icon-question {
  width: 16px;
  height: 16px;
}
.icon-question:hover {
  color: #409eff;
}

.footer {
  font-weight: 500;
  text-align: center;
  padding: 1em;
}

.action-name {
  font-style: normal;
  font-weight: 400;
  font-size: 64px;
  line-height: 80px;
  color: #000000;
}

.domain {
  font-style: normal;
  font-weight: 400;
  font-size: 32px;
  line-height: 40px;
  color: #000000;
}

.detail {
  padding-left: 2em;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
}

.error-info {
  display: flex;
  padding-top: 8px;
}

.error-message {
  font-style: normal;
  font-weight: 400;
  font-size: 18px;
  line-height: 20px;
  color: #888888;
  padding-right: 8px;
}

.action-panel {
  margin-top: 32px;
}
</style>
