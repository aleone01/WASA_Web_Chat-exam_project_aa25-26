<script>
export default {
	props: {
    msg: String
  }
}
</script>

<template>
  <div class="alert alert-danger" role="alert">
    {{ msg }}
  </div>
</template>

<style>
</style>
