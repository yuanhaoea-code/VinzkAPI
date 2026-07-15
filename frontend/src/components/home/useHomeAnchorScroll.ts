export function scrollToHomeSection(hash: string) {
  if (!hash.startsWith('#')) {
    return
  }

  const target = document.querySelector<HTMLElement>(hash)
  if (!target) {
    return
  }

  const scrollToTarget = () => {
    const navHeight = document.querySelector<HTMLElement>('.home-nav')?.getBoundingClientRect().height ?? 0
    const top = target.getBoundingClientRect().top + window.scrollY - navHeight - 18

    window.scrollTo({
      top: Math.max(0, top),
      behavior: 'auto'
    })
  }

  scrollToTarget()
  window.requestAnimationFrame(scrollToTarget)
  window.setTimeout(scrollToTarget, 120)

  window.history.replaceState(null, '', hash)
}
