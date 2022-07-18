<template>
  <vue-form
    v-model="formData"
    :ui-schema="uiSchema"
    :schema="schema"
    :formProps="formProps"
    @submit="submit"
    @cancel="cancel"
  >
  </vue-form>
</template>

<script>
import { getConfigSchema, saveConfig, getConfig } from "@/js/service";
import { defineComponent } from "@vue/runtime-core";

export default defineComponent({
  name: "Setting",
  data() {
    return {
      formData: {},
      defaultData: {},
      schema: {},
      formProps: {},
    };
  },
  mounted() {
    console.log("mounted");
     getConfigSchema().then((data) => {
      console.log("schema", data);
      this.schema = data;
    });
    this.getConfigData();
  },
  methods: {
    submit() {
      console.log("submit");
      saveConfig(this.formData).then(() => {
        this.$notify({
          type: 'success',
          message: '配置更新成功',
        });
        this.getConfigData();
      });
    },
    cancel() {
      console.log("cancel");
      this.formData = this.defaultData;
    },
    getConfigData() {
      getConfig().then((data) => {
        console.log("data", data);
        this.formData = data;
        this.defaultData = data;
      });
    },
  },
});
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
