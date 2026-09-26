import { mount } from 'svelte'
import './style.css'
import App from './App.svelte'

const target = document.getElementById('app')
const app = mount(App, {
  target: target
})

export default app
