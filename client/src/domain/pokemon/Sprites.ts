import { InvalidSpriteUrlError } from '@/domain/errors'

export interface SpritesData {
  frontDefault: string | null
  frontShiny: string | null
  backDefault: string | null
  backShiny: string | null
  officialArtwork?: string | null
}

export class Sprites {
  private readonly _frontDefault: string | null
  private readonly _frontShiny: string | null
  private readonly _backDefault: string | null
  private readonly _backShiny: string | null
  private readonly _officialArtwork: string | null

  constructor(data: SpritesData) {
    this.validateUrl('frontDefault', data.frontDefault)
    this.validateUrl('frontShiny', data.frontShiny)
    this.validateUrl('backDefault', data.backDefault)
    this.validateUrl('backShiny', data.backShiny)
    this.validateUrl('officialArtwork', data.officialArtwork)

    this._frontDefault = data.frontDefault
    this._frontShiny = data.frontShiny
    this._backDefault = data.backDefault
    this._backShiny = data.backShiny
    this._officialArtwork = data.officialArtwork ?? null
  }

  private validateUrl(field: string, url: string | null | undefined): void {
    if (url === null || url === undefined) {
      return
    }

    if (typeof url !== 'string' || url.trim() === '') {
      throw new InvalidSpriteUrlError(`${field}: ${url}`)
    }

    try {
      new URL(url)
    } catch {
      throw new InvalidSpriteUrlError(`${field}: ${url}`)
    }
  }

  get frontDefault(): string | null {
    return this._frontDefault
  }

  get frontShiny(): string | null {
    return this._frontShiny
  }

  get backDefault(): string | null {
    return this._backDefault
  }

  get backShiny(): string | null {
    return this._backShiny
  }

  get officialArtwork(): string | null {
    return this._officialArtwork
  }

  getBestQuality(): string | null {
    return this._officialArtwork ?? this._frontDefault ?? this._frontShiny
  }

  getShinySprite(): string | null {
    return this._frontShiny ?? this._backShiny
  }

  hasAnySprite(): boolean {
    return !!(
      this._frontDefault ||
      this._frontShiny ||
      this._backDefault ||
      this._backShiny ||
      this._officialArtwork
    )
  }
}
