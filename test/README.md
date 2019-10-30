# Testsetup

## Create

```./render.sh | kubectl apply -f -```

## Destroy

```./render.sh | kubectl apply -f -```

## Customize

The ```render.sh``` script uses [spruce](https://github.com/geofffranks/spruce) to render
the indivual kubernetes resources. The script supports a few environment variables to
customize a some values in the resource yamls. Please check the script to see what
env variables are available.

## Development

Each resource should be in a separate ```.yml``` file. Each file that should be rendered
must be added explicitely to the ```render.sh``` script. Ensure proper yaml document
separation by adding ```---``` where appropriate.

All resources must use the ```VSB_DEV_ID``` env variable to construct the resource name and
the ```dev-id``` label. The default in case ```VSB_DEV_ID``` is not set must be "dev".

Image references must be constructed using the ```VSB_DEV_REGISTRY``` env variable and
an empty default.

If in doupt, try to use the existing resource files as examples.
