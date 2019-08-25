import request from '../../api/request'

const system = {
  namespaced: true,
  state: {
    config: {
      enable_register: false,
      installed: false
    }
  },
  getters: {
    config () {
      const config = localStorage.getItem('system_config')
      if (config) {
        return JSON.parse(config).data
      }
      return {}
    }
  },
  mutations: {
    SET_SYSTEM_CONFIG: (state, value) => {
      state.config = value
      console.log(value)
      localStorage.setItem('system_config', JSON.stringify(value))
      return value
    }
  },
  actions: {
    async getSettings ({ state, commit, getters }, inputData = { forceMode: false }) {
      let config = localStorage.getItem('system_config') || '{}'
      config = JSON.parse(config) || {}
      const isForce = inputData.forceMode;

      if (isForce || typeof config.expired_time === 'undefined' || config.expired_time < +(new Date())) {
        const { data } = await request.get('/system/config')
        const cacheLife = data.cache_life > 0 ? data.cache_life : 300

        config = {
          expired_time: +(new Date()) + cacheLife * 1000,
          data: data.data
        }
      } else {
        return config.data
      }
      commit('SET_SYSTEM_CONFIG', config)
      return config.data
    }
  }
}
export default system
