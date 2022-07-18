<template>
  <div>
    <el-form
      ref="ruleFormRef"
      :model="form"
      :rules="rules"
      class="demo-ruleForm"
      status-icon
    >
      <el-form-item label="账号" prop="username">
        <el-input v-model="form.username" placeholder="请输入账号" />
      </el-form-item>
      <el-form-item label="密码" prop="password">
        <el-input
          type="password"
          v-model="form.password"
          placeholder="请输入密码"
          show-password
        />
      </el-form-item>
    </el-form>
    <div class="form-button">
      <el-button type="primary" @click="submitForm()">登录</el-button>
      <el-button @click="resetForm()">重置</el-button>
    </div>
  </div>
</template>

<script>
import { ElMessage } from "element-plus";
export default {
  props: {
    modelValue: Object,
  },
  data() {
    return {
      form: {
        username: "",
        password: "",
      },
      rules: {
        username: [
          { required: true, message: "请输入账号", trigger: "blur" },
          { min: 3, max: 5, message: "长度应为3到5", trigger: "blur" },
        ],
        password: [
          { required: true, message: "请输入密码", trigger: "blur" },
          { min: 5, max: 12, message: "长度应为5到12", trigger: "blur" },
        ],
      },
    }
  },
  watch: {
    form() {
      this.$emit("update:modelValue", this.form);
    },
  },
  mounted() {
    console.log("mounted");
    this.form = this.modelValue;
  },
  methods: {
    // 登录
    submitForm() {
      this.$refs.ruleFormRef.validate((valid, result) => {
        if (!valid) {
          console.log(result);
          ElMessage.error("表单验证失败");
          return;
        }
        console.log(this.ruleForm);
        this.$emit("submitForm", this.form);
      });
    },
    // 重置表单
    resetForm() {
      this.$refs.ruleFormRef.resetFields();
    },
  }
};
</script>

<style scoped>
.demo-ruleForm {
  width: 299px;
  height: 135px;
}
.form-button {
  display: flex;
  justify-content: center;
}
</style>
