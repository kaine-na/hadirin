// Root layout: server-side rendering dimatikan karena stack berbasis SPA
// yang berkomunikasi ke API backend terpisah (tidak perlu SSR).
export const ssr = false;
export const prerender = false;
export const trailingSlash = 'never';
