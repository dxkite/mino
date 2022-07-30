<template>
  <div class="login-content">
    <div class="header">
      <HeaderImg></HeaderImg>
    </div>
    <div class="main">
      <LoginItem v-model="ruleForm" @submitForm="submitForm" />
    </div>
    <Footer />
  </div>
</template>

<script>
import LoginItem from "../components/LoginItem.vue";
import HeaderImg from "../components/HeaderImg.vue";
import Footer from "../components/Footer.vue";
export default {
  name: "login",
  components: {
    LoginItem,
    HeaderImg,
    Footer
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
.login-content {
  display: flex;
  height: 100vh;
  width: 100%;
  flex-direction: column;
  align-items: center;
  justify-content: center;
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
