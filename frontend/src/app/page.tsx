import { Button } from "@/components/ui/button"

export default function Home() {
	return (
		<div className="flex flex-col flex-1 items-center justify-center bg-zinc-50 font-sans dark:bg-black">
			<main className="flex flex-1 w-full max-w-3xl items-center justify-center py-32 px-16 bg-white dark:bg-black gap-8">
				<h1 className="text-center font-heading text-2xl font-bold">Welcome!</h1>

				<Button>Click me</Button>
			</main>
		</div>
	)
}
