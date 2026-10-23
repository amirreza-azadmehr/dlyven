/** Public identity shown in the running app. Keep display strings stable; hrefs are opened in the OS browser. */
export const AUTHOR = {
  name: 'Amirreza Azadmehr',
  links: [
    { id: 'github', href: 'https://github.com/amirreza-azadmehr', display: 'github.com/amirreza-azadmehr' },
    { id: 'linkedin', href: 'https://tr.linkedin.com/in/amirreza-azadmehr', display: 'linkedin.com/in/amirreza-azadmehr' },
    { id: 'telegram', href: 'https://t.me/amir_reza_azadmehr', display: '@amir_reza_azadmehr' },
  ],
} as const

export type AuthorLinkId = (typeof AUTHOR.links)[number]['id']
