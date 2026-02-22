import { computed } from "vue";

export function usePublicHooks() {
  const switchStyle = computed(() => {
    return {
      "--el-switch-on-color": "#409eff",
      "--el-switch-off-color": "#dcdfe6"
    };
  });

  return {
    switchStyle
  };
}

