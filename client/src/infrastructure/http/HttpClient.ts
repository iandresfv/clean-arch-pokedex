/**
 * RFC 9457 problem details, the error shape the Pokedex API returns.
 */
export interface ProblemDetails {
  type: string;
  title: string;
  status: number;
  detail?: string;
  instance?: string;
  requestId?: string;
}

/**
 * Error carrying the server's problem details.
 *
 * `requestId` is the correlation identifier the API attaches to every log line
 * for the request, so a failure reported by a user can be traced to its logs.
 */
export class HttpError extends Error {
  readonly status: number;
  readonly problem: ProblemDetails | null;
  readonly requestId: string | null;

  constructor(message: string, status: number, problem: ProblemDetails | null) {
    super(message);
    this.name = 'HttpError';
    this.status = status;
    this.problem = problem;
    this.requestId = problem?.requestId ?? null;
  }

  get isNotFound(): boolean {
    return this.status === 404;
  }

  get isRateLimited(): boolean {
    return this.status === 429;
  }
}

/**
 * Minimal `fetch` wrapper with a timeout and problem+json decoding.
 *
 * A bare `fetch` has no timeout: a hung connection leaves the promise pending
 * forever and the UI stuck in a loading state with nothing to retry.
 */
export class HttpClient {
  private readonly baseUrl: string;
  private readonly timeoutMs: number;

  constructor(baseUrl: string, timeoutMs: number) {
    this.baseUrl = baseUrl.replace(/\/+$/, '');
    this.timeoutMs = timeoutMs;
  }

  async getJSON<T>(path: string, signal?: AbortSignal): Promise<T> {
    const url = `${this.baseUrl}${path}`;

    // Combining signals lets a caller-driven cancellation (a component
    // unmounting) and the timeout both abort the same request.
    const timeout = AbortSignal.timeout(this.timeoutMs);
    const combined = signal ? AbortSignal.any([signal, timeout]) : timeout;

    let response: Response;
    try {
      response = await fetch(url, {
        signal: combined,
        headers: { Accept: 'application/json' },
      });
    } catch (cause) {
      if (timeout.aborted) {
        throw new HttpError(
          `Request to ${path} timed out after ${String(this.timeoutMs)}ms`,
          0,
          null
        );
      }
      throw cause;
    }

    if (!response.ok) {
      throw new HttpError(
        await describeFailure(response, path),
        response.status,
        await problemOf(response)
      );
    }

    return (await response.json()) as T;
  }
}

/**
 * Reads the problem body without letting a malformed one mask the real status:
 * an error page from a proxy is not JSON, and parsing it must not throw.
 */
async function problemOf(response: Response): Promise<ProblemDetails | null> {
  const contentType = response.headers.get('Content-Type') ?? '';
  if (!contentType.includes('json')) {
    return null;
  }
  try {
    return (await response.clone().json()) as ProblemDetails;
  } catch {
    return null;
  }
}

async function describeFailure(response: Response, path: string): Promise<string> {
  const problem = await problemOf(response);
  if (problem?.title) {
    return problem.detail ? `${problem.title}: ${problem.detail}` : problem.title;
  }
  return `Request to ${path} failed with status ${String(response.status)}`;
}
