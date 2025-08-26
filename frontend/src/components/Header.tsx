import Link from "next/link";

export function Header() {
  return (
    <header className="flex items-center justify-between whitespace-nowrap border-b border-solid border-b-[#ededed] px-6 md:px-10 py-3">
      <div className="flex items-center gap-3 text-[#141414]">
        <div className="size-5 text-[#141414]">
          <svg viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
            <g clipPath="url(#clip0_6_330)">
              <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M24 0.757355L47.2426 24L24 47.2426L0.757355 24L24 0.757355ZM21 35.7574V12.2426L9.24264 24L21 35.7574Z"
                fill="currentColor"
              />
            </g>
            <defs>
              <clipPath id="clip0_6_330"><rect width="48" height="48" fill="white" /></clipPath>
            </defs>
          </svg>
        </div>
        <h2 className="text-[#141414] text-lg font-bold leading-tight tracking-[-0.015em]">TimeTrack</h2>
      </div>
      <div className="hidden md:flex flex-1 justify-end gap-8">
        <nav className="flex items-center gap-6">
          <Link className="text-[#141414] text-sm font-medium leading-normal" href="#">Product</Link>
          <Link className="text-[#141414] text-sm font-medium leading-normal" href="#">Solutions</Link>
          <Link className="text-[#141414] text-sm font-medium leading-normal" href="#">Resources</Link>
          <Link className="text-[#141414] text-sm font-medium leading-normal" href="#">Pricing</Link>
        </nav>
        <div className="flex gap-2">
          <Link
            href="#"
            className="flex min-w-[84px] max-w-[480px] items-center justify-center overflow-hidden rounded-full h-10 px-4 bg-[#999999] text-[#141414] text-sm font-bold leading-normal tracking-[0.015em]"
          >
            <span className="truncate">Get Started</span>
          </Link>
          <Link
            href="#"
            className="flex min-w-[84px] max-w-[480px] items-center justify-center overflow-hidden rounded-full h-10 px-4 bg-[#ededed] text-[#141414] text-sm font-bold leading-normal tracking-[0.015em]"
          >
            <span className="truncate">Contact Sales</span>
          </Link>
        </div>
      </div>
    </header>
  );
}

