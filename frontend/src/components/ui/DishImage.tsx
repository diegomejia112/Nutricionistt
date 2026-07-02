import { useState, useEffect } from 'react'
import { UtensilsCrossed } from 'lucide-react'
import { clsx } from 'clsx'

interface Props {
  id: string
  url?: string
  nombre: string
  categoria: string
  className?: string
}

const CATEGORY_STYLE: Record<string, { bg: string; text: string }> = {
  'Desayuno': { bg: '#FEF9C3', text: '#854D0E' },
  'Comida':   { bg: '#DCFCE7', text: '#166534' },
  'Cena':     { bg: '#EDE9FE', text: '#5B21B6' },
  'Colacion': { bg: '#FFE4E6', text: '#9F1239' },
  'Sopa':     { bg: '#DBEAFE', text: '#1E3A8A' },
  'Guisado':  { bg: '#FEE2E2', text: '#991B1B' },
  'Antojito': { bg: '#FEF3C7', text: '#92400E' },
  'Postre':   { bg: '#FCE7F3', text: '#9D174D' },
  'Bebida':   { bg: '#E0F2FE', text: '#0C4A6E' },
}

function getStyle(categoria: string) {
  return CATEGORY_STYLE[categoria] ?? { bg: '#F3F4F6', text: '#374151' }
}

function Skeleton({ className }: { className?: string }) {
  return (
    <div className={clsx('relative overflow-hidden bg-gray-100', className)}>
      <div className="absolute inset-0 -translate-x-full animate-[shimmer_1.4s_infinite]"
        style={{ background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.6), transparent)' }} />
    </div>
  )
}

function Placeholder({ nombre, categoria, className }: { nombre: string; categoria: string; className?: string }) {
  const { bg, text } = getStyle(categoria)
  return (
    <div className={clsx('flex flex-col items-center justify-center gap-1.5 select-none', className)}
      style={{ background: bg }}>
      <UtensilsCrossed size={20} color={text} strokeWidth={1.5} />
      <span className="text-lg font-bold leading-none" style={{ color: text }}>{nombre.charAt(0).toUpperCase()}</span>
    </div>
  )
}

type Status = 'loading' | 'downloading' | 'ok' | 'error'

export default function DishImage({ id, nombre, categoria, className }: Props) {
  const [status, setStatus] = useState<Status>('loading')
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    setStatus('loading')
    setAttempt(0)
  }, [id])

  useEffect(() => {
    if (status !== 'downloading') return
    const t = setTimeout(() => setAttempt(a => a + 1), 3000)
    return () => clearTimeout(t)
  }, [status, attempt])

  function handleLoad() { setStatus('ok') }

  function handleError() {
    fetch(`/api/imagenes/${id}`)
      .then(r => setStatus(r.status === 202 ? 'downloading' : 'error'))
      .catch(() => setStatus('error'))
  }

  const showSkeleton = status === 'loading' || status === 'downloading'

  return (
    <div className={clsx('relative overflow-hidden rounded-lg', className)}>
      {showSkeleton && <Skeleton className="absolute inset-0 rounded-lg" />}
      {status === 'error' && (
        <Placeholder nombre={nombre} categoria={categoria} className="absolute inset-0 rounded-lg" />
      )}
      {status !== 'error' && (
        <img
          key={attempt}
          src={`/api/imagenes/${id}?a=${attempt}`}
          alt={nombre}
          onLoad={handleLoad}
          onError={handleError}
          className={clsx(
            'w-full h-full object-cover transition-opacity duration-500',
            status === 'ok' ? 'opacity-100' : 'opacity-0 absolute inset-0'
          )}
        />
      )}
    </div>
  )
}
