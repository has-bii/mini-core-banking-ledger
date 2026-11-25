import Image from "next/image"
import { Button } from "@/components/ui/button"

export default function Home() {
	return (
		<div className="flex flex-col flex-1 items-center justify-center bg-zinc-50 font-sans dark:bg-black">
			<main className="flex flex-1 w-full max-w-3xl flex-col items-center justify-center gap-8 bg-white px-6 py-32 sm:px-16 dark:bg-black">
				<Image
					src="/brand/still-bank-horizontal-color.svg"
					alt="Still Bank"
					width={206}
					height={64}
				/>
				<h1 className="text-center font-heading text-2xl font-bold">Welcome!</h1>

				<Button>Click me</Button>
			</main>
		</div>
	)
}
