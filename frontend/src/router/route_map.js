const routes = {
  register_admin: { path: '/signup', component: () => import('@/views/login/index'), hidden: true },
  sign_up: { path: '/signup', component: () => import('@/views/login/index'), hidden: true }

}

export const GetDynamicRoute = (name) => {
  return routes[name]
}
