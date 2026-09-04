<script setup lang="ts">
type PageWidth = 'fluid' | 'wide' | 'form'
type PageDensity = 'default' | 'compact'

withDefaults(defineProps<{
  as?: 'main' | 'section' | 'div'
  width?: PageWidth
  density?: PageDensity
}>(), {
  as: 'main',
  width: 'fluid',
  density: 'default',
})
</script>

<template>
  <component
    :is="as"
    class="admin-ui-page-shell"
    :class="[`admin-ui-page-shell--${width}`, `admin-ui-page-shell--${density}`]"
  >
    <slot />
  </component>
</template>

<style scoped>
.admin-ui-page-shell {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: var(--admin-ui-section-gap);
  width: 100%;
  min-width: 0;
  margin: 0 auto;
  padding: var(--admin-ui-page-gutter);
  color: var(--admin-ui-color-text);
  letter-spacing: 0;
}

.admin-ui-page-shell--wide {
  max-width: var(--admin-ui-page-max-wide);
}

.admin-ui-page-shell--form {
  max-width: var(--admin-ui-page-max-form);
}

.admin-ui-page-shell--compact {
  --admin-ui-section-gap: 12px;
}

@media (max-width: 767px) {
  .admin-ui-page-shell {
    padding: var(--admin-ui-page-gutter-mobile);
  }
}
</style>
