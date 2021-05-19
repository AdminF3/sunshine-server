package cmd

import (
	"context"
	"log"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/seed"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/spf13/cobra"
)

var (
	ctx = context.Background()

	users, orgs, assets, prjs int
)

const longM = `
The 'seed' command creates fake but valid data.

By default it creates one admin, one pd and some users. Number of users can be modified by flag.

The organizations are created only with LEAR. For LEAR is selected one of the users, picked by index of its position of creation (if there are more organizations than users it will cycle again through the users).
-->Nit: if you want one user to have more than one organization, just seed with few users and a lot of organizations. For example 2 users and 10 org will create 2 users being lears to 5 organizations.

Assets are created with owner one organization of the created ones chosen consecutively. Number of assets to create can be modified with flag.

Projects are created with country 'Latvia'; their number can be modified with flag; asset is chosen from a consecutive position of creation. Organization is that of the asset. For every asset is created one project, if there are enough assets.
--> Nit: If you want assets with no projects - number of assets to create should be greater to number of projects.
`

// seedCmd represents the migrate command
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database with fake but valid data.",
	Long:  longM,
	Run:   execSeed,
}

func init() {
	seedCmd.Flags().IntVarP(&users, "users", "u", 5, "Number of users to create")
	seedCmd.Flags().IntVarP(&orgs, "organizations", "o", 2, "Number of organizations to create")
	seedCmd.Flags().IntVarP(&assets, "assets", "a", 2, "Number of assets to create")
	seedCmd.Flags().IntVarP(&prjs, "projects", "p", 2, "Number of projects to create")

	rootCmd.AddCommand(seedCmd)
}

func execSeed(_ *cobra.Command, _ []string) {
	log.SetFlags(0)

	env, err := services.NewEnv()
	if err != nil {
		log.Fatalf("cannot setup environment: %v", err)
	}

	log.Println("Start seeding...")

	s := seed.NewSeeder(env)

	users, err := s.Users(ctx, users)
	if err != nil {
		log.Fatalf("creating users fails: %v", err)
	}
	printUsers(users)

	orgs, err := s.Organizations(ctx, users, orgs)
	if err != nil {
		log.Fatalf("creating organization fails: %v", err)
	}
	printOrgs(orgs)

	assets, err := s.Assets(ctx, orgs, assets)
	if err != nil {
		log.Fatalf("creating assets fails: %v", err)
	}
	log.Printf("\n%d assets created", len(assets))

	n, err := s.Projects(ctx, assets, users["admin"], prjs)
	if err != nil {
		log.Fatalf("creating projects fails: %v", err)
	}
	log.Printf("\n%d projects created", n)

	log.Print("\nDone.")
}

func printUsers(data seed.Users) {
	log.Println("")

	for t, users := range data {
		for user, pass := range users {
			log.Printf("%q - email: %q with password: %q (id: %q)\n", t, user.Email, pass, user.ID)
		}
	}
}

func printOrgs(data []*models.Organization) {
	log.Println("")

	for _, o := range data {
		log.Printf("organization with lear: %q\n", o.Roles.LEAR)
	}
}
