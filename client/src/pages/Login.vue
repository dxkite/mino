<template>
  <div class="login-conent">
    <div class="header">
      <HeaderImg></HeaderImg>
    </div>
    <div class="main">
      <LoginItem v-model="ruleForm" @submitForm="submitForm" />
    </div>
    <div class="footer">mino 网络访问助手 v0.2.6-beta a82db2</div>
  </div>
</template>

<script>
import LoginItem from "../components/LoginItem.vue";
import HeaderImg from "../components/HeaderImg.vue";
export default {
  name: "login",
  components: {
    LoginItem,
    HeaderImg,
  },
  data() {
    return {
      ruleForm: {
        username: "",
        password: "",
      },
    };
  },
  methods: {
    submitForm() {},
  },
};
</script>
<style scoped>
.login-conent {
  display: flex;
  height: 100vh;
  width: 100%;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.footer {
  margin-top: 180px;
  font-family: "Inter";
  font-style: normal;
  font-weight: 400;
  font-size: 14px;
  line-height: 17px;
}
.header {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.main {
  margin-top: 50px;
  width: 299px;
}
</style>
