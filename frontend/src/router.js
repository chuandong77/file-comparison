import { createRouter, createWebHashHistory } from 'vue-router';
import Home from './components/home.vue';
import ComparisonResult from './components/comparison-result.vue';

const routes = [
    {
        path: "/",
        redirect: "/home",
    },
    {
        path: '/home',
        component: Home
    },
    {
        path: '/comparison-result',
        component: ComparisonResult
    }
];

const router = createRouter({
    history: createWebHashHistory(),
    routes
});

export default router;
