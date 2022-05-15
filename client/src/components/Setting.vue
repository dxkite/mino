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
import service from "@/js/service";

export default {
  name: "Setting",
  data() {
    return {
      formData: {},
      defaultData: {},
      schema: {},
      formProps: {},
    };
  },
  created() {
    service.getConfigSchema().then((data) => {
      console.log("schema", data);
      this.schema = data;
    });
  },
  mounted() {
    console.log("mounted");
    this.getConfig();
  },
  methods: {
    submit() {
      console.log("submit");
      service.saveConfig(this.formData).then(() => {
        this.$notify({
          type: 'success',
          message: '配置更新成功',
        });
        this.getConfig();
      });
    },
    cancel() {
      console.log("cancel");
      this.formData = this.defaultData;
    },
    getConfig() {
      service.getConfig().then((data) => {
        console.log("data", data);
        this.formData = data;
        this.defaultData = data;
      });
    },
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
