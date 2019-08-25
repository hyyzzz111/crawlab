
const dynamicRouting = {
  namespaced: true,
  state: {
    routes: []
  },
  actions: {
    pushRoute ({ getters, commit }, params) {
      let routes = getters.routes
      routes.push(params)
      commit('PUSH_ROUTE', routes)
    }
  },
  mutations: {
    PUSH_ROUTE (state, value) {
      state.routes = value
      localStorage.setItem('dynamic_routes', JSON.stringify(value))
    }
  },
  getters: {
    routes (state) {
      if (state.routes.length > 0) {
        return state.routes
      }
      let dynamicRoutes = localStorage.getItem('dynamic_routes')
      if (!dynamicRoutes) {
        dynamicRoutes = []
      } else {
        dynamicRoutes = JSON.parse(dynamicRoutes)
      }
      return dynamicRoutes
    }
  }
}

export default dynamicRouting
